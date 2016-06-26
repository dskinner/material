#version 100
#extension GL_OES_standard_derivatives : enable
#define onesqrt2 0.70710678118
#define sqrt2 1.41421356237
#define pi 3.14159265359
#define twopi 6.28318530718

#define touchBegin 0.0
#define touchMove 1.0
#define touchEnd 2.0
precision mediump float;

// TODO pass this in some other way so sampler can be selected
// 0:fontsize, 1:pad, 2:edge
uniform vec4 glyphconf;

uniform sampler2D texglyph;
uniform sampler2D texicon;
uniform vec2 glyph;
uniform vec4 shadowColor;

varying vec4 vtexcoord;
varying vec4 vvertex;
varying vec4 vtouch;

// interpolated distance, and size values
// x, y is unit value [0..1]
// z, w is material width, height
varying vec4 vdist;
varying vec4 vcolor;

vec4 sampleIcon() {
  vec4 clr = texture2D(texicon, vtexcoord.xy);
  clr.rgb += vcolor.rgb;
  clr.a *= 0.54; // https://www.google.com/design/spec/style/color.html#color-ui-color-application
  return clr;
}

float median(float a, float b, float c) {
  return max(min(a, b), min(max(a, b), c));
}

float hue(vec3 c) {
  float lo = min(c.r, min(c.g, c.b));
  float hi = max(c.r, max(c.g, c.b));

  float n;
  if (c.r == hi) {
    n = (c.g-c.b)/(hi-lo);
  } else if (c.g == hi) {
    n = 2.0 + (c.b-c.r)/(hi-lo);
  } else if (c.b == hi) {
    n = 4.0 + (c.r-c.g)/(hi-lo);
  }
  n *= 60.0;
  if (n < 0.0) {
    n += 360.0;
  }
  return n;
}

vec3 hsl(vec3 c) {
  float lo = min(c.r, min(c.g, c.b));
  float hi = max(c.r, max(c.g, c.b));

  float h = hue(c);

  float sat = 0.0;
  if (lo == hi) {
    return vec3(hue(c), 0.0, 0.0);
  }
  float lum = (lo+hi)/2.0;
  if (lum < 0.5) {
    sat = (hi-lo)/(hi+lo);
  } else {
    sat = (hi-lo)/(2.0-hi-lo);
  }
  return vec3(hue(c), sat, lum);
}

float saturation(vec3 c) {
  float lo = min(c.r, min(c.g, c.b));
  float hi = max(c.r, max(c.g, c.b));

  if (hi != 0.0) {
    return (hi-lo)/hi;
  }
  return 0.0;
}

vec3 hsv2rgb(float h, float s, float v) {
  float c = v*s;
  float x = c * (1.0-abs(mod((h/60.0), 2.0) - 1.0));
  float m = v-c;

  vec3 clr;
  if (0.0 <= h && h < 60.0) {
    clr = vec3(c, x, 0.0);
  } else if (60.0 <= h && h < 120.0) {
    clr = vec3(x, c, 0.0);
  } else if (120.0 <= h && h < 180.0) {
    clr = vec3(0.0, c, x);
  } else if (180.0 <= h && h < 240.0) {
    clr = vec3(0.0, x, c);
  } else if (240.0 <= h && h < 300.0) {
    clr = vec3(x, 0.0, c);
  } else {
    clr = vec3(c, 0.0, x);
  }
  return clr+m;
}

vec3 layerhue(vec3 c) {
  vec3 inv = 1.0-c;

  float h = hue(inv);
  float s = saturation(c);
  float v = max(c.r, max(c.g, c.b));

  // return c;
  return hsv2rgb(h, s, v);

  // return inv;
}

float special(vec3 c) {
  float s = saturation(c);
  float v = max(c.r, max(c.g, c.b));

  // if (v > 0.5) {
    // return v;
  // }
  // if (v > 0.4 && s > 0.1) {
    // return v;
  // }
  // if (v  0.55 && s > 0.1) {
  // return v;
  // }
  // return 0.0;
  // if (s <= 0.05) {
    // return 0.9;
  // }
  return v;
  // return s;
}

vec4 sampleGlyph() {
  float fontsize = glyphconf.x;
  float pad = glyphconf.y;
  float edge = glyphconf.z;

  vec4 s = texture2D(texglyph, vtexcoord.xy);

  // vec4 tmp = (1.0-s);
  // tmp.a = 1.0;
  // tmp.rgb = s.rgb-tmp.rgb;
  // tmp.rgb *= 4.0;

  // return tmp;

  // float m;
  // float v = max(tmp.r, max(tmp.g, tmp.b));
  // tmp.rgb = vec3(1.0);
  // float m = v;
  // return vec4(vec3(1.0),m);
  // return tmp;
  // float m = median(tmp.r, tmp.g, tmp.b);
  // if (m > 0.0) {
    // m = 0.5;
  // }

  // s = 1.0-s;
  // s.a = median(s.r, s.g, s.b);
  // s.a += 0.3;
  // return s;

  // vec3 c = layerhue(s.rgb);
  // return vec4(1.0-c, 1.0);

  float m = median(s.r, s.g, s.b);
  // float m = special(s.rgb);
  // vec3 tmp = hsl(s.rgb);
  // float m = tmp.g-tmp.b;
  // return vec4(m);

  // without derivatives
  // m -= 0.5;
  // vec4 tmp = vec4(0.0);
  // vec2 sz = vec2(1.0/1366.0, 1.0/760.0);
  // vec2 sz = vec2(1.0/vdist.z, 1.0/vdist.w);
  // tmp = texture2D(texglyph, vtexcoord.xy+sz.x);
  // float dfdx = median(tmp.r, tmp.g, tmp.b)-m;
  // tmp = texture2D(texglyph, vtexcoord.xy+sz.y);
  // float dfdy = median(tmp.r, tmp.g, tmp.b)-m;
  // float fw = abs(dfdx)+abs(dfdy);
  // m = clamp(m/fw + 0.5, 0.0, 1.0);

  // with derivatives
  m -= 0.5;
  m = clamp(m/fwidth(m) + 0.5, 0.0, 1.0);

  // float d = texture2D(texglyph, vtexcoord.xy).a;
  // float h = vdist.w;
  // float gamma = 0.22/(pad*(h/fontsize));

  // d += 0.2; // bold
  // d -= 0.2; // thin

  vec4 clr = vcolor;
  // clr.a = smoothstep(edge-gamma, edge+gamma, d);
  // clr.r = smoothstep(edge-gamma, edge+gamma, m);

  // gamma = 0.22;
  // clr.a = smoothstep(edge-gamma, edge+gamma, m);
  // clr = mix(vec4(0.0), vcolor, m);
  clr.a = m;

  // clr.g = 0.0;
  // clr.b = 0.0;
  // clr.rgb = vec3(smoothstep(edge-gamma, edge+gamma, m));
  // clr.rgb = vec3(smoothstep(0.0, 0.5+gamma, m));
  // clr.a *= 0.87; // secondary text
  return clr;
}

// TODO drop this
bool shouldcolor(vec2 pos, float sz) {
  pos = 0.5-abs(pos-0.5);
  pos *= vdist.zw;

  if (pos.x <= sz && pos.y <= sz) {
    float d = length(1.0-(pos/sz));
    if (d > 1.0) {
      return false;
    }
  }
  return true;
}

float shade(vec2 pos, float sz) {
  pos = 0.5-abs(pos-0.5);
  pos *= vdist.zw;

  if (pos.x <= sz && pos.y <= sz) {
    float d = length(1.0-(pos/sz));
    if (d > 1.0) { // TODO consider moving this as discard into top of main
      return 0.0;
    }
    return 1.0-d;
  } else if (pos.x <= sz && pos.y > sz) {
    return pos.x/sz;
  } else if (pos.x > sz && pos.y <= sz) {
    return pos.y/sz;
  }
  return 1.0;
}

void main() {
  float roundness = vvertex.w;

	if (vtexcoord.x >= 0.0) {
    if (vtexcoord.z == 1.0) {
      gl_FragColor = sampleIcon();
    } else {
      gl_FragColor = sampleGlyph();
    }
	} else if (vvertex.z <= 0.0) { // draw shadow
    if (roundness < 8.0) {
      roundness = 8.0;
    }

    // TODO ellipsis are being over-rounded resulting in a distortion
    // in the shadow. The distortion isn't very noticable unless drawing
    // only shadows, but this does affect outline of material making it
    // more difficult to see material edges. Clamping roundness results
    // in a strange distortion when animating roundness and the cause is
    // not currently clear.
    //
    // At the same time, the distortion looks better for different cases.
    roundness += -vvertex.z;
    // if (roundness > vdist.z/2.0) {
    // roundness = vdist.z/2.0;
    // }

		gl_FragColor = shadowColor;

    // maps roundness to 1 - [0..0.66] and helps shadows cast by rectangles and ellipses
    // look similar at the same z index.
    // TODO vvertex.w is roundness, double check usage here as was used before float roundness declared
    float e = 1.0 - (vvertex.w/vdist.z/0.75);

    gl_FragColor.a = smoothstep(0.0, e, shade(vdist.xy, roundness));
    vec2 n = abs(vdist.xy*2.0 - 1.0);
    n = n*n*n*n;
    n = 1.0-n;

    // reduce alpha/strength as z-index increases
    float f = 1.0 + (-vvertex.z*0.1);

    gl_FragColor.a *= n.x*n.y/f;
    // gl_FragColor.a *= 10.0;
  } else { // draw material
    if (shouldcolor(vdist.xy, roundness)) {
      gl_FragColor = vcolor;

      // anti-alias roundness
      if (gl_FragColor.a != 0.0) { // guard against exposing bg of text and icon content
        float dist = 1.0-shade(vdist.xy, roundness);
        // fractional based on largest size, approximates a consistent value across resolutions
        float dt = (5.0/max(vdist.z, vdist.w));
        gl_FragColor.a = 1.0-smoothstep(1.0-dt, 1.0, dist);
      }

      // respond to touch with radial
      const float dur = 200.0;
      const float clr = 0.035;

      float since = vtouch.w;
      vec4 react = vec4(0.0);

      // float d = length(vtouch.xy-vdist.xy);
      // float d = length(vtouch.xx-vdist.xx); // horizontal sweep

      const float magic = 100.0;
      float d = length((vtouch.xy*vdist.zw)-(vdist.xy*vdist.zw));
      d /= magic;

      float t = since/dur;
      if (d < 2.0*t) {
        if (t < sqrt2) { // color in
          react = vec4(clr);
        } else if (t < pi) { // fade out
          float fac = 1.0 - (t-sqrt2)/(pi-sqrt2);
          react = vec4(fac*clr);
        }
      }

      gl_FragColor += react;
    } else {
      discard;
    }
	}
}
