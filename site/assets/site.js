// host-wails project page. Taken from the spartan-native site (site/assets/site.js): nav, reveal on
// scroll, sub nav, syntax highlighting, copy buttons and the footer year. No dependencies.

document.documentElement.classList.remove('no-js');

const $ = (sel, root = document) => root.querySelector(sel);
const $$ = (sel, root = document) => [...root.querySelectorAll(sel)];
const reduceMotion = matchMedia('(prefers-reduced-motion: reduce)').matches;

// ------------------------------------------------------------------ Nav

const nav = $('.nav');
$('.nav-toggle')?.addEventListener('click', () => {
  const open = nav.toggleAttribute('data-open');
  $('.nav-toggle').setAttribute('aria-expanded', String(open));
});

// ------------------------------------------------------------------ Reveal on scroll

const revealer = new IntersectionObserver(
  (entries) => {
    for (const e of entries) {
      if (e.isIntersecting) {
        e.target.classList.add('in');
        revealer.unobserve(e.target);
      }
    }
  },
  { rootMargin: '0px 0px -10% 0px' },
);
$$('.reveal, .shot.tilt').forEach((el) => revealer.observe(el));

// Sub nav: highlight the section in view
const subLinks = $$('.subnav a');
if (subLinks.length) {
  const byId = new Map(subLinks.map((a) => [a.getAttribute('href').slice(1), a]));
  const spy = new IntersectionObserver(
    (entries) => {
      for (const e of entries) {
        if (e.isIntersecting) {
          subLinks.forEach((a) => a.removeAttribute('data-active'));
          byId.get(e.target.id)?.setAttribute('data-active', '');
        }
      }
    },
    { rootMargin: '-40% 0px -55% 0px' },
  );
  byId.forEach((_, id) => {
    const el = document.getElementById(id);
    if (el) spy.observe(el);
  });
}

// ------------------------------------------------------------------ Syntax highlighting

const esc = (s) => s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');

const KEYWORDS =
  'import|from|export|const|let|var|return|new|interface|type|readonly|extends|await|async|function|if|else|true|false|null|undefined|satisfies|default|of|for|track';

const GRAMMARS = {
  ts: [
    ['c', /\/\/[^\n]*|\/\*[\s\S]*?\*\//y],
    ['s', /'[^'\n]*'|"[^"\n]*"|`[^`]*`/y],
    ['k', new RegExp(`\\b(?:${KEYWORDS})\\b`, 'y')],
    ['n', /\b\d[\d.]*\b/y],
    ['f', /\b[A-Za-z_$][\w$]*(?=\s*[(<])/y],
    ['t', /\b[A-Z][\w]*\b/y],
  ],
  scss: [
    ['c', /\/\/[^\n]*|\/\*[\s\S]*?\*\//y],
    ['s', /'[^'\n]*'|"[^"\n]*"/y],
    ['k', /@[\w-]+/y],
    ['a', /--[\w-]+/y],
    ['n', /#[0-9a-fA-F]{3,8}\b|\b\d[\d.]*(?:px|rem|%)?/y],
    ['f', /\b[\w-]+(?=\()/y],
  ],
  html: [
    ['c', /<!--[\s\S]*?-->/y],
    ['s', /"[^"]*"/y],
    ['t', /<\/?[\w-]+|\/?>/y],
    ['a', /[\w\-[\]()*.#@]+(?==)/y],
    ['k', /@(?:for|if|else)\b/y],
  ],
  bash: [
    ['c', /#[^\n]*/y],
    ['s', /'[^'\n]*'|"[^"\n]*"/y],
    ['f', /(?<=^|\n|&& )(?:npm|npx|git|node|go|make|cd)\b/y],
  ],
  go: [
    ['c', /\/\/[^\n]*|\/\*[\s\S]*?\*\//y],
    ['s', /"[^"\n]*"|`[^`]*`/y],
    ['k', /\b(?:package|import|func|return|var|const|type|struct|interface|map|range|for|if|else|go|defer|nil|true|false)\b/y],
    ['n', /\b\d[\d.]*\b/y],
    ['f', /\b[A-Za-z_][\w]*(?=\s*\()/y],
    ['t', /\b[A-Z][\w]*\b/y],
  ],
};
GRAMMARS.json = GRAMMARS.ts;

function highlight(code, lang) {
  const rules = GRAMMARS[lang];
  if (!rules) return esc(code);
  let out = '';
  let plain = '';
  let i = 0;
  outer: while (i < code.length) {
    for (const [cls, re] of rules) {
      re.lastIndex = i;
      const m = re.exec(code);
      if (m && m[0].length) {
        out += esc(plain) + `<span class="tok-${cls}">${esc(m[0])}</span>`;
        plain = '';
        i += m[0].length;
        continue outer;
      }
    }
    plain += code[i++];
  }
  return out + esc(plain);
}

$$('pre code[data-lang]').forEach((el) => {
  el.innerHTML = highlight(el.textContent, el.dataset.lang);
});

// ------------------------------------------------------------------ Copy buttons

const COPY_ICON =
  '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><rect x="9" y="9" width="12" height="12" rx="2"/><path d="M5 15V5a2 2 0 0 1 2-2h10"/></svg>';
const DONE_ICON =
  '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M5 12.5l4.5 4.5L19 7"/></svg>';

$$('[data-copy]').forEach((btn) => {
  btn.innerHTML = COPY_ICON;
  btn.addEventListener('click', async () => {
    const target = btn.dataset.copy ? document.getElementById(btn.dataset.copy) : null;
    const text = target ? target.textContent : btn.closest('.code')?.querySelector('pre:not([hidden])')?.textContent;
    if (!text) return;
    try {
      await navigator.clipboard.writeText(text.trim());
      btn.innerHTML = DONE_ICON;
      btn.setAttribute('data-copied', '');
      setTimeout(() => {
        btn.innerHTML = COPY_ICON;
        btn.removeAttribute('data-copied');
      }, 1600);
    } catch {
      /* clipboard blocked: nothing to do */
    }
  });
});

// ------------------------------------------------------------------ Tabs (code blocks)

$$('[role="tablist"]').forEach((list) => {
  const tabs = $$('[role="tab"]', list);
  const select = (tab) => {
    tabs.forEach((t) => {
      const on = t === tab;
      t.setAttribute('aria-selected', String(on));
      t.tabIndex = on ? 0 : -1;
      const panel = document.getElementById(t.getAttribute('aria-controls'));
      if (panel) panel.hidden = !on;
    });
  };
  tabs.forEach((t, i) => {
    t.addEventListener('click', () => select(t));
    t.addEventListener('keydown', (e) => {
      const d = e.key === 'ArrowRight' ? 1 : e.key === 'ArrowLeft' ? -1 : 0;
      if (!d) return;
      const next = tabs[(i + d + tabs.length) % tabs.length];
      next.focus();
      select(next);
    });
  });
});


// ------------------------------------------------------------------ Footer year

$$('[data-year]').forEach((el) => (el.textContent = new Date().getFullYear()));
