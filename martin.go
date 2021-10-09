/*
 * Copyright 2018 Johannes Donath <johannesd@torchmind.com>
 * and other copyright owners as documented in the project's IP log.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package sstv

import (
	"errors"
	"image"

	"github.com/go-audio/audio"
)

type MartinMode uint8

const (
	Martin1 MartinMode = 44
	Martin2 MartinMode = 40
)

const (
	martin1PulseLength = .4576 //协议指定的每个亮度值信号持续的时长 0.4576ms
	martin2PulseLength = .2288

	martinLineFrequency = 1200
	martinLineLength    = 4.862

	martinSeparatorLength    = .572
	martinSeparatorFrequency = 1500 //用1500 ~ 2300Hz的信号来表示不同亮度值
)

// 提供Martin Emmerson设计的Martin实现

// 此实现在114或58秒内将RGB图像编码为240行
// M1:114s	 M2:58s
type martinEncoder struct {
	mode   MartinMode
	format *audio.Format
}

// 创建与Martin兼容的新图像编码器
func NewMartin(mode MartinMode, format *audio.Format) Encoder {
	return &martinEncoder{
		mode:   mode,
		format: format,
	}
}

func (enc *martinEncoder) Vis() uint8 {
	return uint8(enc.mode)
}

func (enc *martinEncoder) Resolution() image.Rectangle {
	return image.Rect(0, 0, 320, 256) //M1模式下，图片大小为320*256
}

func (enc *martinEncoder) Encode(img image.Image) *audio.FloatBuffer {
	var pulseLength float64
	switch enc.mode {
	case Martin1:
		pulseLength = martin1PulseLength
	case Martin2:
		pulseLength = martin2PulseLength
	default:
		panic(errors.New("illegal encoding mode"))
	}

	wr := newWriter(enc.format)
	wr.writeHeader()
	wr.writeVis(uint8(enc.mode))

	size := img.Bounds().Size()
	for y := 0; y < size.Y; y++ {
		wr.write(martinLineFrequency, martinLineLength)

		for i := 0; i < 3; i++ {
			wr.write(martinSeparatorFrequency, martinSeparatorLength)

			for x := 0; x < size.X; x++ {
				r, g, b := convertRGB(img.At(x, y))

				var val float64
				switch i {
				case 0:
					val = g
				case 1:
					val = b
				case 2:
					val = r
				}
				wr.writeValue(val, pulseLength)
			}

			wr.write(martinSeparatorFrequency, martinSeparatorLength)
		}
	}

	return wr.buf
}
