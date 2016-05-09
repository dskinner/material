#version 100

attribute vec4 vertex;
attribute vec4 color;

// sample with xy, offset at zw
attribute vec4 texcoord;

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

void main() {
	gl_Position = vec4(vertex.xyz, 1.0) * view * proj;
	vcolor = color;
	vtexcoord = texcoord;
	vvertex = vertex;
	vdist = dist;
}
