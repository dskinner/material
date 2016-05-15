#version 100

attribute vec4 vertex;
attribute vec4 color;

// sample with xy, offset at zw
attribute vec4 texcoord;

// xy is relative position of originating touch event
// z is state of touch event; begin (0), move (1), end (2)
// w is timing information
attribute vec4 touch;

// v0, v1 == 1 && v2, v3 == 0, interpolated in fragment shader
// to determine location since all materials are drawn as single
// mesh.
attribute vec4 dist;

// uniform mat4 world;
uniform mat4 view;
uniform mat4 proj;

varying vec4 vcolor;
varying vec4 vtexcoord;
varying vec4 vdist;
varying vec4 vvertex;
varying vec4 vtouch;

void main() {
	// TODO review if fragment shader *really* needs access to z coord
	// of shadow's material.
	vec4 vert = vec4(vertex.xyz, 1.0);
	if (vert.z < 0.0) {
		vert.z = 0.0;
	}
	// gl_Position = vert * view * proj;
  gl_Position = proj * view * vert;
	vcolor = color;
	vtexcoord = texcoord;
	vvertex = vertex;
	vdist = dist;
	vtouch = touch;
}
