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

	"golang.org/x/mobile/exp/gl/glutil"
	"golang.org/x/mobile/gl"
) // }}}

// {{{ vertex data
var tableVertices = f32.Bytes(binary.LittleEndian,
	// triangle 1
	-.5, -.5,
	.5, .5,
	-.5, .5,

	// triangle 2
	-.5, -.5,
	.5, -.5,
	.5, .5,

	// Line1
	-.5, 0,
	.5, 0,

	// mallets
	0, -.25,
	0, .25,
)

// }}}

// {{{ shader
var vShader = `#version 100
attribute vec4 a_Position;

void main() {
	gl_Position = a_Position;
	gl_PointSize = 10.0;
}
`

var fShader = `#version 100
precision mediump float;

uniform vec4 u_Color;

void main() {
	gl_FragColor = u_Color;
}
` // }}}

// {{{ global value
var (
	buf      gl.Buffer
	color    gl.Uniform
	position gl.Attrib
	program  gl.Program
) // }}}

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
	color = glctx.GetUniformLocation(program, "u_Color")

}

func onStop(glctx gl.Context) {
	glctx.DeleteProgram(program)
	glctx.DeleteBuffer(buf)
}

var green float32

func onPaint(ctx gl.Context, sz size.Event) {
	ctx.ClearColor(1, 0, 0, 1)
	ctx.Clear(gl.COLOR_BUFFER_BIT)

	ctx.UseProgram(program)
	// ctx.Enable(gl

	// setting uniform
	green += 0.01
	if green >= 1 {
		green = 0
	}
	ctx.Uniform4f(color, 0, green, 0, 1)

	// // buffer settings
	ctx.BindBuffer(gl.ARRAY_BUFFER, buf)
	ctx.EnableVertexAttribArray(position)
	ctx.VertexAttribPointer(position, 2, gl.FLOAT, false, 0, 0)

	// ctx.DrawArrays(gl.TRIANGLES, 0, int(len(tableVertices)))
	ctx.DrawArrays(gl.TRIANGLES, 0, 6)

	// draw Line
	ctx.Uniform4f(color, 1, 1, 1, 1)
	ctx.DrawArrays(gl.LINES, 6, 2)

	// draw mallets
	// GL_VERTEX_PROGRAM_POINT_SIZE is not define in go.
	// https://forums.khronos.org/showthread.php/5984-gl_PointSize-problem
	ctx.Enable(0x8642)
	ctx.Uniform4f(color, 1, 0, 0, 1)
	ctx.DrawArrays(gl.POINTS, 8, 1)
	ctx.Uniform4f(color, 0, 0, 1, 1)
	ctx.DrawArrays(gl.POINTS, 9, 1)

	ctx.DisableVertexAttribArray(position)

} // }}}
