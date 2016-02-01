#version 100

attribute vec4 position;
attribute vec2 tc0;

uniform mat4 world;
uniform mat4 view;
uniform mat4 proj;

varying float vheight;
varying vec3 vpos;
varying vec2 vtc0;

void main() {
	vec4 pos = position;
	gl_Position = pos * world * view * proj;
	vpos = pos.xyz;
	vtc0 = tc0;
  vheight = world[1][1];
}
