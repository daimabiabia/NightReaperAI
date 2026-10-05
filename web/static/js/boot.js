/**
 * NightReaper AI — 开机序列动画 (by yyyr)
 * 每个浏览器会话只播放一次；prefers-reduced-motion 用户自动跳过。
 * 自包含：动态创建 DOM/样式，不影响页面逻辑。
 */
(function () {
  try {
    if (sessionStorage.getItem('nrai_boot_v1')) return;
    if (window.matchMedia && window.matchMedia('(prefers-reduced-motion: reduce)').matches) {
      sessionStorage.setItem('nrai_boot_v1', '1');
      return;
    }
    sessionStorage.setItem('nrai_boot_v1', '1');
  } catch (e) { /* storage 不可用时仍播放一次 */ }

  // ?theme=dark / ?theme=light 直达链接支持
  try {
    var q = new URLSearchParams(location.search);
    var tq = q.get('theme');
    if (tq === 'dark' || tq === 'light') {
      document.documentElement.setAttribute('data-theme', tq);
      document.documentElement.setAttribute('data-theme-preference', tq);
      localStorage.setItem('nightreaper-theme', tq);
    }
  } catch (e) { /* noop */ }

  // ---- 样式 ----
  var css = [
    '#nrai-boot{position:fixed;inset:0;z-index:99999;background:#04070a;',
    'display:flex;flex-direction:column;align-items:center;justify-content:center;',
    'font-family:ui-monospace,Consolas,"Courier New",monospace;overflow:hidden;',
    'transition:opacity .6s ease;cursor:default;user-select:none}',
    '#nrai-boot.nrai-out{opacity:0;pointer-events:none}',
    '#nrai-boot canvas{position:absolute;inset:0;width:100%;height:100%;opacity:0;transition:opacity .8s ease}',
    '#nrai-boot canvas.nrai-on{opacity:.5}',
    '.nrai-core{position:relative;z-index:2;text-align:center;pointer-events:none}',
    '.nrai-word{font-size:clamp(28px,6vw,64px);font-weight:700;letter-spacing:.22em;color:#00ff88;',
    'text-shadow:0 0 12px rgba(0,255,136,.85),0 0 42px rgba(0,255,136,.45),0 0 90px rgba(0,255,136,.25);',
    'opacity:0;transform:translateY(10px);transition:opacity .9s ease,transform .9s ease;margin:0 0 14px}',
    '.nrai-word.nrai-on{opacity:1;transform:translateY(0)}',
    '.nrai-sub{font-size:clamp(11px,1.4vw,14px);letter-spacing:.34em;color:#39d98a;min-height:1.6em;',
    'opacity:.85;text-shadow:0 0 8px rgba(0,255,136,.5)}',
    '.nrai-sub .nrai-caret{display:inline-block;width:.6em;color:#00ff88;animation:nrai-blink .7s steps(1) infinite}',
    '@keyframes nrai-blink{50%{opacity:0}}',
    '.nrai-barwrap{width:min(340px,62vw);height:3px;background:rgba(0,255,136,.14);margin:26px auto 0;border-radius:2px;overflow:hidden}',
    '.nrai-bar{height:100%;width:0;background:linear-gradient(90deg,#00b877,#00ff88);',
    'box-shadow:0 0 10px rgba(0,255,136,.8);transition:width .18s linear}',
    '.nrai-granted{margin-top:20px;font-size:clamp(12px,1.6vw,16px);letter-spacing:.4em;color:#7dffb8;',
    'opacity:0;font-weight:700}',
    '.nrai-granted.nrai-on{opacity:1;animation:nrai-flash .42s steps(2) 3}',
    '@keyframes nrai-flash{50%{opacity:.15}}',
    '.nrai-foot{position:absolute;bottom:22px;left:0;right:0;text-align:center;font-size:10px;',
    'letter-spacing:.3em;color:rgba(0,255,136,.4);z-index:2}'
  ].join('');

  var style = document.createElement('style');
  style.textContent = css;
  document.head.appendChild(style);

  var root = document.createElement('div');
  root.id = 'nrai-boot';
  root.innerHTML =
    '<canvas></canvas>' +
    '<div class="nrai-core">' +
    '<h1 class="nrai-word">NIGHTREAPER</h1>' +
    '<div class="nrai-sub"><span class="nrai-text"></span><span class="nrai-caret">█</span></div>' +
    '<div class="nrai-barwrap"><div class="nrai-bar"></div></div>' +
    '<div class="nrai-granted">ACCESS GRANTED</div>' +
    '</div>' +
    '<div class="nrai-foot">NIGHTREAPER AI v1.0.0 · BY YYYR · 女娲安全工作室</div>';
  document.body.appendChild(root);

  // ---- 矩阵雨 ----
  var canvas = root.querySelector('canvas');
  var ctx = canvas.getContext('2d');
  var GLYPHS = 'アカサタナハマヤラワ01<>;$#{}[]/*+= NightReaper';
  var cols = 0, drops = [], FS = 14;
  function resize() {
    canvas.width = window.innerWidth;
    canvas.height = window.innerHeight;
    cols = Math.ceil(canvas.width / FS);
    drops = new Array(cols).fill(0).map(function () { return Math.random() * -60; });
  }
  resize();
  window.addEventListener('resize', resize);

  var rainTimer = setInterval(function () {
    ctx.fillStyle = 'rgba(4,7,10,0.10)';
    ctx.fillRect(0, 0, canvas.width, canvas.height);
    ctx.font = FS + 'px ui-monospace, Consolas, monospace';
    for (var i = 0; i < cols; i++) {
      var ch = GLYPHS.charAt(Math.floor(Math.random() * GLYPHS.length));
      var y = drops[i] * FS;
      // 首字符亮白绿，拖尾深绿
      ctx.fillStyle = Math.random() < 0.06 ? '#d8ffe9' : '#00c26e';
      ctx.fillText(ch, i * FS, y);
      if (y > canvas.height && Math.random() > 0.972) drops[i] = 0;
      drops[i]++;
    }
  }, 55);

  // ---- 序列 ----
  var sub = root.querySelector('.nrai-text');
  var bar = root.querySelector('.nrai-bar');
  var word = root.querySelector('.nrai-word');
  var granted = root.querySelector('.nrai-granted');
  var SUB_TEXT = 'REAPING THE DARK // OPERATOR: YYYR';

  canvas.classList.add('nrai-on');
  setTimeout(function () { word.classList.add('nrai-on'); }, 420);

  // 打字机
  var ti = 0;
  var typeTimer = setInterval(function () {
    sub.textContent = SUB_TEXT.slice(0, ++ti);
    if (ti >= SUB_TEXT.length) clearInterval(typeTimer);
  }, 42);

  // 进度条
  var prog = 0;
  var barTimer = setInterval(function () {
    prog = Math.min(100, prog + Math.random() * 9 + 3);
    bar.style.width = prog + '%';
    if (prog >= 100) clearInterval(barTimer);
  }, 130);

  setTimeout(function () { granted.classList.add('nrai-on'); }, 3300);
  setTimeout(function () {
    root.classList.add('nrai-out');
    setTimeout(function () {
      clearInterval(rainTimer);
      if (root.parentNode) root.parentNode.removeChild(root);
      style.parentNode && style.parentNode.removeChild(style);
    }, 700);
  }, 4100);

  // 点击/按键跳过
  function skip() {
    root.classList.add('nrai-out');
    setTimeout(function () {
      clearInterval(rainTimer); clearInterval(typeTimer); clearInterval(barTimer);
      if (root.parentNode) root.parentNode.removeChild(root);
    }, 350);
  }
  root.addEventListener('click', skip);
  window.addEventListener('keydown', skip, { once: true });
})();

/* ---- 页面标题光标彩蛋：NIGHTREAPER AI ▌ 闪烁 ---- */
(function () {
  var base = 'NIGHTREAPER AI';
  var on = true;
  setInterval(function () {
    on = !on;
    document.title = base + (on ? ' ▌' : '  ');
  }, 900);
})();
