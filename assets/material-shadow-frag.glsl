#version 100
#define onesqrt2 0.70710678118
#define sqrt2 1.41421356237
precision mediump float;

uniform vec4 color;

varying vec2 vpos;
varying float vwz;

float alpha = 1.0; // alpha independent of actual constraint size

float csumx(vec2 pos, float lbndl, float lbndu, float ubndl, float ubndu) {
	if (pos.y < lbndl || pos.y > ubndu) {
		return 0.0;
	} else if (lbndl <= pos.x && pos.x <= lbndu) {
		return alpha*(pos.x/lbndu);
	} else if (lbndu < pos.x && pos.x < ubndl) {
		return alpha;
	} else if (ubndl <= pos.x && pos.x <= ubndu) {
		return alpha*((1.0-pos.x)/(1.0-ubndl));
	}
	return 0.0;
}

float csumy(vec2 pos, float lbndl, float lbndu, float ubndl, float ubndu) {
	if (pos.x < lbndl || pos.x > ubndu) {
		return 0.0;
	} else if (lbndl <= pos.y && pos.y <= lbndu) {
		float a = alpha*(pos.y/lbndu);
		float b = csumx(pos, lbndl, lbndu, ubndl, ubndu);
		return a*b;
	} else if (lbndu < pos.y && pos.y < ubndl) {
		float a = alpha;
		float b = csumx(pos, lbndl, lbndu, ubndl, ubndu);
		return a*b;
	} else if (ubndl <= pos.y && pos.y <= ubndu) {
		float a = alpha*((1.0-pos.y)/(1.0-ubndl));
		float b = csumx(pos, lbndl, lbndu, ubndl, ubndu);
		return a*b;
	}
	return 0.0;
}

float csum(vec2 pos, float wz) {	
	float r = sqrt2-1.0;
	float zstep = smoothstep(8.0, 1.0, 1.0+vwz); 
	float z = r*zstep;
	vec2 norm = pos*2.0-1.0;
	// sentinel for mid
	float bndstp = smoothstep(0.0, 7.0, vwz);
	float bnd = mix(0.10, 0.24, bndstp);
	float rbnd = mix(0.42, 0.31, bndstp);
	float y0;
	if (bnd <= vpos.x && vpos.x <= 1.0-bnd && bnd <= vpos.y && vpos.y <= 1.0-bnd && length(vpos-0.5) < rbnd) {
		y0 = alpha;
	} else {
		norm = mix(norm*norm*norm*norm, norm*norm, smoothstep(0.0, 7.0, vwz));
		// scale down if necessary
		norm *= 2.0-smoothstep(0.0, 7.0, vwz); // stay in-sync with mix(y1, y0, ...) at return of func
		// in addition to above, provides smooth corner transition
		float t = onesqrt2 + z; // factor to smooth right angle (norm*onesqrt2) to round corner (norm)
		norm = vec2(length(norm*t)); // square to rounded corner with next return
		norm = t-norm; // flip axis x, y; t factor is my max x,y
		y0 = csumy(norm, 0.0, 0.5-z, 0.5+z, 1.0); // wip
	}
	float za = 0.25*smoothstep(0.0, 11.0, vwz);
	vec2 ypos = pos;//*(2.0-smoothstep(0.0, 7.0, vwz));
	float y1 = csumy(ypos, 0.0, 0.1+za, 0.9-za, 1.0);
	if (vpos.x <= 0.1 || 0.9 <= vpos.x || vpos.y <= 0.1 || 0.9 <= vpos.y) { // trim
		y1 *= smoothstep(0.0, 9.0, vwz);
	}
	return mix(y1, y0, smoothstep(0.0, 9.0, vwz));
}

void main() {
	gl_FragColor = color;
	float realalpha = 0.4;
	// vec2 pos = vec2(length((vpos-0.5)*onesqrt2));
	gl_FragColor.a = csum(vpos, vwz)*realalpha;
}
