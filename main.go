package main

// {{{ import
import (
	"encoding/binary"
	"log"
	"math"

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
	// Order of XYRGB
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
	projectionMat4 *f32.Mat4 = new(f32.Mat4)
	modelMat4      *f32.Mat4 = new(f32.Mat4)
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
				onSizeChanged(glctx, sz)
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

func onSizeChanged(ctx gl.Context, sz size.Event) {
	if ctx != nil {
		//ctx.Viewport(0, 0, sz.WidthPx, sz.HeightPx)
	}
	// initialize motrix
	p := new(f32.Mat4)
	p.Identity()
	modelMat4.Identity()
	modelMat4.Translate(p, 0, 0, -2.5)
	log.Printf("translate:\n%v", modelMat4) // 想定通りの値
	// RotateXM(modelMat4, DegreeToRadian(60))
	// p.Identity()
	modelMat4 = RotateXM(modelMat4, DegreeToRadian(30))
	log.Printf("rotate:\n%v", p)
	// rotate * translate
	//modelMat4 = Mul(p, modelMat4)
	//log.Printf("rotate * translate:\n%v", p)

	if projectionMat4 != nil {
		// var aspectRatio float32
		// if sz.WidthPx > sz.HeightPx {
		// 	aspectRatio = float32(sz.WidthPx / sz.HeightPx)
		// 	Ortho(projectionMat4, -aspectRatio, aspectRatio, -1, 1, -1, 1)
		// } else {
		// 	aspectRatio = float32(sz.HeightPx / sz.WidthPx)
		// 	Ortho(projectionMat4, -1, 1, -aspectRatio, aspectRatio, -1, 1)
		// }
		projectionMat4.Identity()
		PerspectiveM(projectionMat4, 45, float32(sz.WidthPx)/float32(sz.HeightPx), 1, 10)
	}

	// calculate projection * model
	projectionMat4 = Mul(projectionMat4, modelMat4)
	log.Println(projectionMat4)
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
	// githubのtutoralだとmat4から[]float32への変換を自分でやっているだと。。。
	ctx.UniformMatrix4fv(projection, ConvertMat4ToFloat32Array(projectionMat4))

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

func Ortho(m *f32.Mat4, left, right, bottom, top, near, far float32) {
	m[0][0] = 2 / (right - left)
	m[0][1] = 0
	m[0][2] = 0

	m[1][0] = 0
	m[1][1] = 2 / (top - bottom)
	m[1][2] = 0

	m[2][0] = 0
	m[2][1] = 0
	m[2][2] = -2 / (far - near)

	m[0][3] = -(right + left) / (right - left)
	m[1][3] = -(top + bottom) / (top - bottom)
	m[2][3] = -(far + near) / (far - near)
	m[3][3] = 1
}

// f32.Mat4 to [16]float32
func ConvertMat4ToFloat32Array(m *f32.Mat4) []float32 {
	out := make([]float32, 16)
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			out[i*4+j] = m[i][j]
		}
	}
	return out
}

func DegreeToRadian(deg float32) f32.Radian {
	return f32.Radian(deg * math.Pi / 180)
}

func CopyMat4(m *f32.Mat4) *f32.Mat4 {
	ret := new(f32.Mat4)
	ret[0][0] = m[0][0]
	ret[0][1] = m[0][1]
	ret[0][2] = m[0][2]
	ret[0][3] = m[0][3]
	ret[1][0] = m[1][0]
	ret[1][1] = m[1][1]
	ret[1][2] = m[1][2]
	ret[1][3] = m[1][3]
	ret[2][0] = m[2][0]
	ret[2][1] = m[2][1]
	ret[2][2] = m[2][2]
	ret[2][3] = m[2][3]
	ret[3][0] = m[3][0]
	ret[3][1] = m[3][1]
	ret[3][2] = m[3][2]
	ret[3][3] = m[3][3]
	return ret
}

func PerspectiveM(m *f32.Mat4, yFovInDegrees, aspect, n, f float32) {
	rad := DegreeToRadian(yFovInDegrees)
	a := float32(1.0 / math.Tan(float64(rad)/2.0))
	m[0][0] = a / aspect
	m[0][1] = 0
	m[0][2] = 0
	m[0][3] = 0

	m[1][0] = 0
	m[1][1] = a
	m[1][2] = 0
	m[1][3] = 0

	m[2][0] = 0
	m[2][1] = 0
	m[2][2] = -((f + n) / (f - n))
	m[2][3] = -1

	m[3][0] = 0
	m[3][1] = 0
	m[3][2] = -((2 * f * n) / (f - n))
	m[3][3] = 0
}

// rotate matrix
func RotateXM(m *f32.Mat4, rad f32.Radian) *f32.Mat4 {
	ret := new(f32.Mat4)
	ret.Identity()
	a := new(f32.Mat4)
	a.Identity()
	r := float64(rad)
	a[1][1] = float32(math.Cos(r))
	a[1][2] = float32(-math.Sin(r))
	a[2][1] = float32(math.Sin(r))
	a[2][2] = float32(math.Cos(r))
	ret.Mul(a, m)
	return ret
}

func Mul(a *f32.Mat4, b *f32.Mat4) *f32.Mat4 {
	ret := new(f32.Mat4)
	ret.Identity()
	ret.Mul(a, b)
	return ret
}

// }}}
