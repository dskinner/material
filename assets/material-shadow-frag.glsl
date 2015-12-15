#version 100
#define onesqrt2 0.70710678118
precision mediump float;

uniform vec4 color;

varying vec2 vpos;
varying float vwz;
varying vec2 sdt;

float csumx(vec2 pos, vec2 dt) {
	if (pos.x < dt.x) {
		return pos.x/dt.x;
	} else if (pos.x > 1.0-dt.x) {
		return (1.0-pos.x)/dt.x;
	}
	return 1.0;
}

float csumy(vec2 pos, vec2 dt) {
	if (pos.y < dt.y) {
		float a = pos.y/dt.y;
		float b = csumx(pos, dt);
		return a*b;
	} else if (pos.y > 1.0-dt.y) {
		float a = (1.0-pos.y)/dt.y;
		float b = csumx(pos, dt);
		return a*b;
	}
	return csumx(pos, dt);
}

float csum() {
	// expand outer blur
	float y0 = csumy(vpos, sdt);

	// expand outer and push mid in based on z
	float y1 = csumy(vpos, sdt/smoothstep(-40.0, 0.0, -clamp(vwz, 0.0, 20.0)));

	// cheap squircle
	vec2 norm = vpos*2.0 - 1.0;
	vec2 pos = norm*norm*norm*norm;
	pos = onesqrt2-vec2(pos*onesqrt2);
	vec2 dt = sdt+0.75;
	float y2 = csumy(pos, dt);

	float a = mix(y2, y0, smoothstep(0.0, 9.0, vwz));
	float b = mix(y1, y2, smoothstep(0.0, 9.0, vwz));
	return a*b;
}

void main() {
	gl_FragColor = color;
	float alpha = smoothstep(-40.0, 0.0, -clamp(vwz, 0.0, 40.0));
	gl_FragColor.a = csum()*clamp(alpha, 0.1, 1.0);
}
