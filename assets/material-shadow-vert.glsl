#version 100

attribute vec4 position;

uniform mat4 world;
uniform mat4 view;
uniform mat4 proj;

varying vec2 vpos;
varying float vwz;

void main() {
	mat4 w = world;
	w[2][3] = 0.0;
	gl_Position = position * w * view * proj;
	vwz = world[2][3];
	vpos = position.xy;
}
