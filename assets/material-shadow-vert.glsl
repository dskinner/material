#version 100

attribute vec4 position;

uniform mat4 world;
uniform mat4 view;
uniform mat4 proj;
uniform vec2 size;

varying vec2 vpos;
varying float vwz;
varying vec2 sdt;

void main() {
	mat4 w = world;
	w[2][3] = 0.0;
	gl_Position = position * w * view * proj;
	vwz = world[2][3];
	vpos = position.xy;
	sdt = size/vec2(world[0][0], world[1][1]);
	sdt = (1.0-sdt)/2.0;
}
