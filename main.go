package main

// {{{ import
import (
	"encoding/binary"
	"log"

	"golang.org/x/mobile/app"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"
	"golang.org/x/mobile/exp/f32"

	"github.com/go-gl/mathgl/mgl32"
	"golang.org/x/mobile/exp/gl/glutil"
	"golang.org/x/mobile/gl"
) // }}}

// {{{ vertex data
var tableVertices = f32.Bytes(binary.LittleEndian,
	// Order of XYZWRGB
	// triangle Fan
	0, 0, 1, 1, 1,
	-.5, -.8, .7, .7, .7,
	.5, -.8, .7, .7, .7,
	.5, .8, .7, .7, .7,
	-.5, .8, .7, .7, .7,
	-.5, -.8, .7, .7, .7,

	// Line1
	-.5, 0, 1, 0, 0,
	.5, 0, 1, 0, 0,

	// mallets
	0, -.4, 0, 0, 1,
	0, .4, 1, 0, 0,
)

// }}}

// {{{ shader
var vShader = `#version 100
uniform mat4 u_Matrix;

attribute vec4 a_Position;
attribute vec4 a_Color;

varying vec4 v_Color;

void main() {
	gl_Position = u_Matrix * a_Position;
	gl_PointSize = 10.0;
	v_Color = a_Color;
}
`

var fShader = `#version 100
precision mediump float;

varying vec4 v_Color;

void main() {
	gl_FragColor = v_Color;
}
` // }}}

// {{{ global value
var (
	buf            gl.Buffer
	color          gl.Attrib
	position       gl.Attrib
	program        gl.Program
	projection     gl.Uniform
	projectionMat4 mgl32.Mat4
) // }}}

// {{{ const

const (
	POSITION_COMPONENT_COUNT = 2
	COLOR_COMPONENT_COUNT    = 3
	BYTE_PER_FLOAT           = 4
	START_COLOR_OFFSET       = POSITION_COMPONENT_COUNT * BYTE_PER_FLOAT
	STRIDE                   = (POSITION_COMPONENT_COUNT + COLOR_COMPONENT_COUNT) * BYTE_PER_FLOAT
)

// }}}

// {{{ main method
func main() {
	app.Main(func(a app.App) {
		var glctx gl.Context
		var sz size.Event
		for e := range a.Events() {
			switch e := a.Filter(e).(type) {
			case lifecycle.Event:
				switch e.Crosses(lifecycle.StageVisible) {
				case lifecycle.CrossOn:
					glctx, _ = e.DrawContext.(gl.Context)
					onStart(glctx)
					a.Send(paint.Event{})
				case lifecycle.CrossOff:
					onStop(glctx)
					glctx = nil
				}
			case size.Event:
				sz = e
				onSizeChanged(sz)
			case paint.Event:
				if glctx == nil || e.External {
					continue
				}
				onPaint(glctx, sz)
				a.Publish()
				a.Send(paint.Event{})
			}
		}
	})
}

// }}}

// {{{ event mehtod
func onStart(glctx gl.Context) {
	log.Print(gl.Version())
	var err error
	// create glProgram
	program, err = glutil.CreateProgram(glctx, vShader, fShader)
	if err != nil {
		log.Printf("error creating GL program: %v", err)
		return
	}

	// buffer settins
	buf = glctx.CreateBuffer()
	glctx.BindBuffer(gl.ARRAY_BUFFER, buf)
	glctx.BufferData(gl.ARRAY_BUFFER, tableVertices, gl.STATIC_DRAW)

	// attribute, uniform settings
	position = glctx.GetAttribLocation(program, "a_Position")
	color = glctx.GetAttribLocation(program, "a_Color")
	projection = glctx.GetUniformLocation(program, "u_Matrix")

}

func onStop(glctx gl.Context) {
	glctx.DeleteProgram(program)
	glctx.DeleteBuffer(buf)
}

func onSizeChanged(sz size.Event) {
	var aspectRatio float32
	if sz.WidthPx > sz.HeightPx {
		aspectRatio = float32(sz.WidthPx / sz.HeightPx)
		//projectionMat4 = mgl32.Ortho(-1, 1, -aspectRatio, aspectRatio, -1, 10)
	} else {
		aspectRatio = float32(sz.HeightPx / sz.WidthPx)
		//projectionMat4 = mgl32.Ortho(-1, 1, -aspectRatio, aspectRatio, -1, 10)
	}

	projectionMat4 = mgl32.Perspective(45, aspectRatio, 1, 10)

	modelMat4 := mgl32.Ident4()
	modelMat4 = modelMat4.Mul4(mgl32.Translate3D(0, 0, -2.5))
	modelMat4 = modelMat4.Mul4(mgl32.HomogRotate3DX(-60))
	projectionMat4 = projectionMat4.Mul4(modelMat4)
}

func onPaint(ctx gl.Context, sz size.Event) {
	ctx.ClearColor(0, 0, 0, 0)
	ctx.Clear(gl.COLOR_BUFFER_BIT)

	ctx.UseProgram(program)

	// // buffer settings
	ctx.BindBuffer(gl.ARRAY_BUFFER, buf)
	// stride, offset is byte number.
	ctx.VertexAttribPointer(position, POSITION_COMPONENT_COUNT, gl.FLOAT, false, STRIDE, 0)
	ctx.EnableVertexAttribArray(position)
	defer ctx.DisableVertexAttribArray(position)
	ctx.VertexAttribPointer(color, COLOR_COMPONENT_COUNT, gl.FLOAT, false, STRIDE, START_COLOR_OFFSET)
	ctx.EnableVertexAttribArray(color)
	defer ctx.DisableVertexAttribArray(color)

	// uniform
	ctx.UniformMatrix4fv(projection, ConvertMat4ToFloat32Array(&projectionMat4))

	ctx.DrawArrays(gl.TRIANGLE_FAN, 0, 6)

	ctx.DrawArrays(gl.LINES, 6, 2)

	// draw mallets
	// GL_VERTEX_PROGRAM_POINT_SIZE is not define in go.
	// https://forums.khronos.org/showthread.php/5984-gl_PointSize-problem
	ctx.Enable(0x8642)
	ctx.DrawArrays(gl.POINTS, 8, 1)
	ctx.DrawArrays(gl.POINTS, 9, 1)

} // }}}

// {{{ f32.Mat4 Helper mehtod

func ConvertMat4ToFloat32Array(m *mgl32.Mat4) []float32 {
	out := make([]float32, 16)
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			out[i*4+j] = m[i*4+j]
		}
	}
	return out
}

// }}}
