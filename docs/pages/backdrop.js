// The landing page's backdrop: pane outlines and box-drawing glyphs that drift
// up in three layers, each shifted by scroll and pointer as far as it is near.
(function () {
  var canvas = document.getElementById("backdrop");
  if (!canvas || !canvas.getContext) return;
  var ctx = canvas.getContext("2d");
  var still = matchMedia("(prefers-reduced-motion: reduce)");
  var glyphs = ["┌", "┐", "└", "┘", "─", "│", "├", "┤", "▣", "▦", "▢", "█", "▸"];
  var depths = [0.25, 0.55, 1];
  var items = [], w = 0, h = 0, dpr = 1;
  var px = 0, py = 0, tx = 0, ty = 0, last = 0;

  function colours() {
    var css = getComputedStyle(document.documentElement);
    return { accent: css.getPropertyValue("--accent").trim(), line: css.getPropertyValue("--dim").trim() };
  }
  var ink = colours();

  function seed() {
    items = [];
    var n = Math.round(Math.min(70, (w * h) / 22000));
    for (var i = 0; i < n; i++) {
      var d = depths[i % depths.length];
      var pane = Math.random() < 0.45;
      items.push({
        x: Math.random() * w,
        y: Math.random() * h,
        d: d,
        pane: pane,
        pw: (40 + Math.random() * 120) * d,
        ph: (30 + Math.random() * 80) * d,
        split: pane && Math.random() < 0.4,
        glyph: glyphs[(Math.random() * glyphs.length) | 0],
        size: 10 + 16 * d,
        speed: 4 + 10 * d,
        accent: Math.random() < 0.3,
        spin: (Math.random() - 0.5) * 0.002
      });
    }
  }

  function resize() {
    dpr = Math.min(devicePixelRatio || 1, 2);
    w = innerWidth; h = innerHeight;
    canvas.width = w * dpr; canvas.height = h * dpr;
    canvas.style.width = w + "px"; canvas.style.height = h + "px";
    ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
    seed();
  }

  function roundRect(x, y, rw, rh, r) {
    ctx.beginPath();
    ctx.moveTo(x + r, y);
    ctx.arcTo(x + rw, y, x + rw, y + rh, r);
    ctx.arcTo(x + rw, y + rh, x, y + rh, r);
    ctx.arcTo(x, y + rh, x, y, r);
    ctx.arcTo(x, y, x + rw, y, r);
    ctx.closePath();
  }

  function draw(dt) {
    ctx.clearRect(0, 0, w, h);
    px += (tx - px) * 0.05; py += (ty - py) * 0.05;
    var scroll = scrollY;
    for (var i = 0; i < items.length; i++) {
      var it = items[i];
      it.y -= it.speed * dt;
      var y = it.y - scroll * it.d * 0.35;
      var span = h + 200;
      y = ((y % span) + span) % span - 100;
      var x = it.x + px * it.d * 18;
      y += py * it.d * 12;
      ctx.globalAlpha = 0.08 + 0.14 * it.d;
      ctx.strokeStyle = ctx.fillStyle = it.accent ? ink.accent : ink.line;
      if (it.pane) {
        ctx.lineWidth = 1 + it.d;
        roundRect(x, y, it.pw, it.ph, 6 * it.d + 2);
        ctx.stroke();
        if (it.split) {
          ctx.beginPath();
          ctx.moveTo(x + it.pw * 0.38, y);
          ctx.lineTo(x + it.pw * 0.38, y + it.ph);
          ctx.stroke();
        }
      } else {
        ctx.font = it.size + "px 'JetBrains Mono', monospace";
        ctx.fillText(it.glyph, x, y);
      }
    }
    ctx.globalAlpha = 1;
  }

  function frame(now) {
    var dt = last ? Math.min((now - last) / 1000, 0.05) : 0;
    last = now;
    draw(dt);
    if (!still.matches && !document.hidden) requestAnimationFrame(frame);
  }

  function start() {
    last = 0;
    if (still.matches) draw(0); else requestAnimationFrame(frame);
  }

  addEventListener("resize", function () { resize(); if (still.matches) draw(0); });
  addEventListener("pointermove", function (e) {
    tx = e.clientX / w - 0.5; ty = e.clientY / h - 0.5;
  });
  addEventListener("scroll", function () { if (still.matches) draw(0); }, { passive: true });
  document.addEventListener("visibilitychange", function () { if (!document.hidden) start(); });
  still.addEventListener("change", start);
  matchMedia("(prefers-color-scheme: light)").addEventListener("change", function () { ink = colours(); });

  resize();
  start();
})();
