package web

const indexView = `<!DOCTYPE html>
<html lang='en'>
<head>
    <meta charset='UTF-8'>
    <meta name='viewport' content='width=device-width, initial-scale=1.0'>
    <title>Flaggo — Feature Flag Manager</title>
    <link href='https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;500;600;700&family=DM+Sans:wght@400;500;600;700&display=swap' rel='stylesheet'>
    <script src='https://unpkg.com/lucide@latest'></script>
    <link href='https://cdnjs.cloudflare.com/ajax/libs/toastr.js/latest/toastr.min.css' rel='stylesheet'>
    <style>
        *, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }

        :root {
            --bg: #0c0e12;
            --bg-raised: #13161c;
            --bg-overlay: #191d25;
            --border: #252a35;
            --border-hover: #3a4150;
            --text: #e4e7ed;
            --text-secondary: #7c8494;
            --text-muted: #4e5668;
            --accent: #6ee7b7;
            --accent-dim: rgba(110, 231, 183, 0.12);
            --accent-border: rgba(110, 231, 183, 0.25);
            --danger: #f87171;
            --danger-dim: rgba(248, 113, 113, 0.12);
            --warning: #fbbf24;
            --warning-dim: rgba(251, 191, 36, 0.12);
            --blue: #60a5fa;
            --blue-dim: rgba(96, 165, 250, 0.12);
            --purple: #a78bfa;
            --purple-dim: rgba(167, 139, 250, 0.12);
            --radius: 8px;
            --radius-lg: 12px;
            --font-mono: 'JetBrains Mono', monospace;
            --font-sans: 'DM Sans', sans-serif;
        }

        body {
            font-family: var(--font-sans);
            background: var(--bg);
            color: var(--text);
            min-height: 100vh;
            -webkit-font-smoothing: antialiased;
        }

        .shell {
            max-width: 960px;
            margin: 0 auto;
            padding: 3rem 1.5rem;
        }

        /* Header */
        .header {
            display: flex;
            align-items: flex-end;
            justify-content: space-between;
            margin-bottom: 2.5rem;
            padding-bottom: 2rem;
            border-bottom: 1px solid var(--border);
        }
        .header-brand {
            font-family: var(--font-mono);
            font-size: 1.5rem;
            font-weight: 700;
            letter-spacing: -0.03em;
            color: var(--accent);
        }
        .header-sub {
            font-size: 0.8125rem;
            color: var(--text-secondary);
            margin-top: 0.25rem;
        }

        /* Stats */
        .stats {
            display: grid;
            grid-template-columns: repeat(3, 1fr);
            gap: 1rem;
            margin-bottom: 2rem;
        }
        .stat-card {
            background: var(--bg-raised);
            border: 1px solid var(--border);
            border-radius: var(--radius-lg);
            padding: 1.25rem;
        }
        .stat-label {
            font-size: 0.6875rem;
            font-weight: 600;
            text-transform: uppercase;
            letter-spacing: 0.08em;
            color: var(--text-muted);
            margin-bottom: 0.5rem;
        }
        .stat-value {
            font-family: var(--font-mono);
            font-size: 1.75rem;
            font-weight: 700;
            color: var(--text);
        }

        /* Buttons */
        .btn {
            display: inline-flex;
            align-items: center;
            justify-content: center;
            gap: 0.5rem;
            font-family: var(--font-sans);
            font-size: 0.8125rem;
            font-weight: 600;
            border: none;
            border-radius: var(--radius);
            cursor: pointer;
            padding: 0.5rem 1rem;
            transition: all 0.15s ease;
            white-space: nowrap;
        }
        .btn-primary {
            background: var(--accent);
            color: var(--bg);
        }
        .btn-primary:hover { filter: brightness(1.1); }
        .btn-ghost {
            background: transparent;
            color: var(--text-secondary);
            border: 1px solid var(--border);
        }
        .btn-ghost:hover {
            border-color: var(--border-hover);
            color: var(--text);
            background: var(--bg-overlay);
        }
        .btn-danger {
            background: var(--danger);
            color: #fff;
        }
        .btn-danger:hover { filter: brightness(1.1); }
        .btn-icon {
            width: 2rem;
            height: 2rem;
            padding: 0;
            background: transparent;
            border: none;
            color: var(--text-muted);
            border-radius: var(--radius);
            cursor: pointer;
            display: inline-flex;
            align-items: center;
            justify-content: center;
            transition: all 0.15s ease;
        }
        .btn-icon:hover {
            background: var(--bg-overlay);
            color: var(--text);
        }

        /* Flag list */
        .flag-list {
            display: flex;
            flex-direction: column;
            gap: 0.75rem;
        }
        .flag-card {
            background: var(--bg-raised);
            border: 1px solid var(--border);
            border-radius: var(--radius-lg);
            padding: 1.25rem;
            transition: border-color 0.15s ease;
        }
        .flag-card:hover {
            border-color: var(--border-hover);
        }
        .flag-top {
            display: flex;
            align-items: center;
            justify-content: space-between;
            gap: 1rem;
        }
        .flag-name {
            font-family: var(--font-mono);
            font-size: 0.9375rem;
            font-weight: 600;
            color: var(--text);
        }
        .flag-actions {
            display: flex;
            align-items: center;
            gap: 0.25rem;
        }
        .flag-meta {
            display: flex;
            align-items: center;
            gap: 0.75rem;
            margin-top: 0.875rem;
            flex-wrap: wrap;
        }

        /* Badges / pills */
        .pill {
            display: inline-flex;
            align-items: center;
            gap: 0.375rem;
            font-size: 0.6875rem;
            font-weight: 600;
            padding: 0.25rem 0.625rem;
            border-radius: 9999px;
            letter-spacing: 0.02em;
        }
        .pill-enabled {
            background: var(--accent-dim);
            color: var(--accent);
            border: 1px solid var(--accent-border);
        }
        .pill-disabled {
            background: rgba(78, 86, 104, 0.2);
            color: var(--text-muted);
            border: 1px solid rgba(78, 86, 104, 0.3);
        }
        .pill-rollout {
            background: var(--blue-dim);
            color: var(--blue);
            border: 1px solid rgba(96, 165, 250, 0.25);
        }
        .pill-actor {
            background: var(--purple-dim);
            color: var(--purple);
            border: 1px solid rgba(167, 139, 250, 0.25);
            font-family: var(--font-mono);
            font-size: 0.625rem;
        }

        /* Toggle switch */
        .toggle {
            position: relative;
            width: 40px;
            height: 22px;
            cursor: pointer;
        }
        .toggle input { display: none; }
        .toggle-track {
            position: absolute;
            inset: 0;
            background: var(--border);
            border-radius: 9999px;
            transition: background 0.2s ease;
        }
        .toggle input:checked + .toggle-track {
            background: var(--accent);
        }
        .toggle-thumb {
            position: absolute;
            top: 3px;
            left: 3px;
            width: 16px;
            height: 16px;
            background: var(--bg);
            border-radius: 9999px;
            transition: transform 0.2s ease;
            box-shadow: 0 1px 3px rgba(0,0,0,0.3);
        }
        .toggle input:checked ~ .toggle-thumb {
            transform: translateX(18px);
        }

        /* Rollout bar */
        .rollout-bar {
            width: 80px;
            height: 4px;
            background: var(--border);
            border-radius: 9999px;
            overflow: hidden;
        }
        .rollout-fill {
            height: 100%;
            background: var(--blue);
            border-radius: 9999px;
            transition: width 0.3s ease;
        }

        /* Dialog */
        .dialog-backdrop {
            position: fixed;
            inset: 0;
            background: rgba(0, 0, 0, 0.6);
            backdrop-filter: blur(4px);
            z-index: 100;
            opacity: 0;
            pointer-events: none;
            transition: opacity 0.2s ease;
        }
        .dialog-backdrop.open {
            opacity: 1;
            pointer-events: auto;
        }
        .dialog {
            position: fixed;
            top: 50%;
            left: 50%;
            transform: translate(-50%, -46%) scale(0.96);
            background: var(--bg-raised);
            border: 1px solid var(--border);
            border-radius: var(--radius-lg);
            padding: 1.75rem;
            width: 90vw;
            max-width: 520px;
            max-height: 85vh;
            overflow-y: auto;
            z-index: 101;
            opacity: 0;
            pointer-events: none;
            transition: all 0.25s cubic-bezier(0.16, 1, 0.3, 1);
        }
        .dialog.open {
            opacity: 1;
            pointer-events: auto;
            transform: translate(-50%, -50%) scale(1);
        }
        .dialog-title {
            font-family: var(--font-mono);
            font-size: 1rem;
            font-weight: 700;
            margin-bottom: 0.25rem;
        }
        .dialog-desc {
            font-size: 0.8125rem;
            color: var(--text-secondary);
            margin-bottom: 1.5rem;
        }
        .dialog-footer {
            display: flex;
            justify-content: flex-end;
            gap: 0.5rem;
            margin-top: 1.5rem;
        }

        /* Form elements */
        .field { margin-bottom: 1.25rem; }
        .field:last-child { margin-bottom: 0; }
        .field-label {
            display: block;
            font-size: 0.6875rem;
            font-weight: 600;
            text-transform: uppercase;
            letter-spacing: 0.08em;
            color: var(--text-secondary);
            margin-bottom: 0.5rem;
        }
        .field-input {
            width: 100%;
            height: 2.5rem;
            padding: 0 0.75rem;
            background: var(--bg);
            border: 1px solid var(--border);
            border-radius: var(--radius);
            color: var(--text);
            font-family: var(--font-mono);
            font-size: 0.8125rem;
            outline: none;
            transition: border-color 0.15s ease;
        }
        .field-input:focus {
            border-color: var(--accent);
        }
        .field-input::placeholder {
            color: var(--text-muted);
        }
        .field-hint {
            font-size: 0.6875rem;
            color: var(--text-muted);
            margin-top: 0.375rem;
        }

        /* Range / slider */
        .range-row {
            display: flex;
            align-items: center;
            gap: 0.75rem;
        }
        .range-slider {
            flex: 1;
            -webkit-appearance: none;
            appearance: none;
            height: 4px;
            background: var(--border);
            border-radius: 9999px;
            outline: none;
        }
        .range-slider::-webkit-slider-thumb {
            -webkit-appearance: none;
            width: 16px;
            height: 16px;
            border-radius: 50%;
            background: var(--accent);
            cursor: pointer;
            box-shadow: 0 0 0 3px var(--accent-dim);
        }
        .range-slider::-moz-range-thumb {
            width: 16px;
            height: 16px;
            border-radius: 50%;
            background: var(--accent);
            cursor: pointer;
            border: none;
        }
        .range-value {
            font-family: var(--font-mono);
            font-size: 0.875rem;
            font-weight: 600;
            color: var(--blue);
            min-width: 3rem;
            text-align: right;
        }

        /* Actor rules section */
        .actor-rules { display: flex; flex-direction: column; gap: 0.5rem; }
        .actor-rule-row {
            display: flex;
            align-items: center;
            gap: 0.5rem;
        }
        .actor-rule-row .field-input { height: 2.25rem; font-size: 0.75rem; }
        .actor-rule-key { width: 140px; flex-shrink: 0; }
        .actor-rule-values { flex: 1; }
        .actor-add-btn {
            font-size: 0.75rem;
            padding: 0.375rem 0.75rem;
            color: var(--accent);
            background: var(--accent-dim);
            border: 1px dashed var(--accent-border);
            border-radius: var(--radius);
            cursor: pointer;
            font-family: var(--font-sans);
            font-weight: 600;
            transition: all 0.15s ease;
            margin-top: 0.25rem;
        }
        .actor-add-btn:hover {
            background: rgba(110, 231, 183, 0.18);
        }

        /* Empty state */
        .empty-state {
            text-align: center;
            padding: 3rem 1rem;
            color: var(--text-muted);
            font-size: 0.875rem;
        }

        /* Scrollbar */
        ::-webkit-scrollbar { width: 6px; }
        ::-webkit-scrollbar-track { background: transparent; }
        ::-webkit-scrollbar-thumb { background: var(--border); border-radius: 9999px; }

        /* Light theme */
        [data-theme='light'] {
            --bg: #f8f9fb;
            --bg-raised: #ffffff;
            --bg-overlay: #f0f1f4;
            --border: #dfe2e8;
            --border-hover: #c4c9d4;
            --text: #1a1d24;
            --text-secondary: #5c6370;
            --text-muted: #9098a6;
            --accent: #059669;
            --accent-dim: rgba(5, 150, 105, 0.1);
            --accent-border: rgba(5, 150, 105, 0.25);
            --danger: #dc2626;
            --danger-dim: rgba(220, 38, 38, 0.08);
            --warning: #d97706;
            --warning-dim: rgba(217, 119, 6, 0.08);
            --blue: #2563eb;
            --blue-dim: rgba(37, 99, 235, 0.08);
            --purple: #7c3aed;
            --purple-dim: rgba(124, 58, 237, 0.08);
        }
        [data-theme='light'] .toggle-thumb {
            background: #fff;
            box-shadow: 0 1px 3px rgba(0,0,0,0.15);
        }
        [data-theme='light'] .pill-disabled {
            background: rgba(0, 0, 0, 0.05);
            color: var(--text-muted);
            border: 1px solid rgba(0, 0, 0, 0.1);
        }
        [data-theme='light'] .pill-actor {
            border-color: rgba(124, 58, 237, 0.2);
        }
        [data-theme='light'] .pill-rollout {
            border-color: rgba(37, 99, 235, 0.2);
        }
        [data-theme='light'] .pill-enabled {
            border-color: rgba(5, 150, 105, 0.2);
        }
        [data-theme='light'] .dialog-backdrop {
            background: rgba(0, 0, 0, 0.3);
        }
        [data-theme='light'] .stat-card,
        [data-theme='light'] .flag-card,
        [data-theme='light'] .dialog {
            box-shadow: 0 1px 3px rgba(0, 0, 0, 0.06);
        }

        /* Checkbox */
        .checkbox-row {
            display: flex;
            align-items: center;
            gap: 0.5rem;
            margin-bottom: 0.625rem;
        }
        .checkbox-row input[type='checkbox'] {
            width: 1rem;
            height: 1rem;
            accent-color: var(--accent);
            cursor: pointer;
        }
        .checkbox-row label {
            font-size: 0.8125rem;
            font-weight: 500;
            color: var(--text);
            cursor: pointer;
        }
        .range-row.disabled {
            opacity: 0.35;
            pointer-events: none;
        }

        /* Theme toggle */
        .theme-toggle {
            width: 2.25rem;
            height: 2.25rem;
            padding: 0;
            background: var(--bg-overlay);
            border: 1px solid var(--border);
            color: var(--text-secondary);
            border-radius: var(--radius);
            cursor: pointer;
            display: inline-flex;
            align-items: center;
            justify-content: center;
            transition: all 0.15s ease;
        }
        .theme-toggle:hover {
            border-color: var(--border-hover);
            color: var(--text);
        }

        @media (max-width: 640px) {
            .stats { grid-template-columns: 1fr; }
            .shell { padding: 1.5rem 1rem; }
        }
    </style>
    <script>
        (function() {
            var t = localStorage.getItem('flaggo-theme');
            if (!t) t = window.matchMedia('(prefers-color-scheme: light)').matches ? 'light' : 'dark';
            document.documentElement.setAttribute('data-theme', t);
        })();
    </script>
</head>
<body>
<div class='shell'>
    <div class='header'>
        <div>
            <div class='header-brand'>flaggo</div>
            <div class='header-sub'>Feature flag management</div>
        </div>
        <div style='display:flex;align-items:center;gap:0.5rem;'>
            <button id='themeToggle' class='theme-toggle' title='Toggle theme'>
                <i data-lucide='sun' class='w-4 h-4' id='themeIconLight' style='display:none;'></i>
                <i data-lucide='moon' class='w-4 h-4' id='themeIconDark' style='display:none;'></i>
            </button>
            <button id='addFlagTrigger' class='btn btn-primary'>
                <i data-lucide='plus' class='w-4 h-4'></i>
                New flag
            </button>
        </div>
    </div>

    <div class='stats'>
        <div class='stat-card'>
            <div class='stat-label'>Total</div>
            <div class='stat-value' id='totalFlagsCount'>0</div>
        </div>
        <div class='stat-card'>
            <div class='stat-label'>Active</div>
            <div class='stat-value' id='activeFlagsCount' style='color: var(--accent);'>0</div>
        </div>
        <div class='stat-card'>
            <div class='stat-label'>Coverage</div>
            <div class='stat-value' id='coveragePercentage'>0%</div>
        </div>
    </div>

    <div style='margin-bottom:0.75rem;position:relative;'>
        <i data-lucide='search' style='position:absolute;left:0.75rem;top:50%;transform:translateY(-50%);width:16px;height:16px;color:var(--text-muted);pointer-events:none;'></i>
        <input type='text' id='searchInput' class='field-input' placeholder='Search flags...' style='padding-left:2.25rem;' />
    </div>

    <div id='featureFlagsContainer' class='flag-list'>
        <div class='empty-state' id='loadingMessage'>Loading flags...</div>
    </div>
</div>

<!-- Create Flag Dialog -->
<div id='createDialogBackdrop' class='dialog-backdrop'></div>
<div id='createDialog' class='dialog'>
    <div class='dialog-title'>New feature flag</div>
    <div class='dialog-desc'>Create a flag with targeting rules and rollout percentage.</div>
    <div class='field'>
        <label class='field-label'>Flag name</label>
        <input type='text' id='createFlagName' class='field-input' placeholder='e.g. dark-mode' />
    </div>
    <div class='field'>
        <label class='field-label'>Rollout percentage</label>
        <div class='checkbox-row'>
            <input type='checkbox' id='createEnableAll' checked />
            <label for='createEnableAll'>Enable all</label>
        </div>
        <div class='range-row disabled' id='createRangeRow'>
            <input type='range' id='createRollout' class='range-slider' min='0' max='100' value='100' />
            <span class='range-value' id='createRolloutValue'>100%</span>
        </div>
        <div class='field-hint'>Percentage of sessions that will see this flag enabled</div>
    </div>
    <div class='field'>
        <label class='field-label'>Actor targeting (optional)</label>
        <div id='createActorRules' class='actor-rules'></div>
        <button type='button' class='actor-add-btn' id='createAddActorBtn'>+ Add rule</button>
        <div class='field-hint'>Targeted actors always see the flag, regardless of rollout %</div>
    </div>
    <div class='dialog-footer'>
        <button class='btn btn-ghost' id='createCancelBtn'>Cancel</button>
        <button class='btn btn-primary' id='createSubmitBtn'>Create</button>
    </div>
</div>

<!-- Edit Flag Dialog -->
<div id='editDialogBackdrop' class='dialog-backdrop'></div>
<div id='editDialog' class='dialog'>
    <div class='dialog-title' id='editDialogTitle'>Edit flag</div>
    <div class='dialog-desc'>Update targeting rules and rollout percentage.</div>
    <input type='hidden' id='editFlagKey' />
    <div class='field'>
        <label class='field-label'>Rollout percentage</label>
        <div class='checkbox-row'>
            <input type='checkbox' id='editEnableAll' />
            <label for='editEnableAll'>Enable all</label>
        </div>
        <div class='range-row' id='editRangeRow'>
            <input type='range' id='editRollout' class='range-slider' min='0' max='100' value='100' />
            <span class='range-value' id='editRolloutValue'>100%</span>
        </div>
    </div>
    <div class='field'>
        <label class='field-label'>Actor targeting</label>
        <div id='editActorRules' class='actor-rules'></div>
        <button type='button' class='actor-add-btn' id='editAddActorBtn'>+ Add rule</button>
    </div>
    <div class='dialog-footer'>
        <button class='btn btn-ghost' id='editCancelBtn'>Cancel</button>
        <button class='btn btn-primary' id='editSubmitBtn'>Save</button>
    </div>
</div>

<!-- Delete Confirm Dialog -->
<div id='deleteDialogBackdrop' class='dialog-backdrop'></div>
<div id='deleteDialog' class='dialog'>
    <div class='dialog-title'>Delete flag</div>
    <div class='dialog-desc'>Are you sure you want to delete <strong id='deleteFlagName'></strong>? This cannot be undone.</div>
    <div class='dialog-footer'>
        <button class='btn btn-ghost' id='deleteCancelBtn'>Cancel</button>
        <button class='btn btn-danger' id='deleteConfirmBtn'>Delete</button>
    </div>
</div>

<script src='https://cdnjs.cloudflare.com/ajax/libs/jquery/3.6.0/jquery.min.js'></script>
<script src='https://cdnjs.cloudflare.com/ajax/libs/toastr.js/latest/toastr.min.js'></script>
<script>
  toastr.options = {
    closeButton: true, progressBar: true, positionClass: 'toast-top-right',
    showDuration: '200', hideDuration: '500', timeOut: '4000',
    showMethod: 'fadeIn', hideMethod: 'fadeOut'
  };

  let featureFlags = {};
  let flagToDelete = '';

  // --- Helpers ---
  function openDialog(id) {
    document.getElementById(id + 'Backdrop').classList.add('open');
    document.getElementById(id).classList.add('open');
  }
  function closeDialog(id) {
    document.getElementById(id + 'Backdrop').classList.remove('open');
    document.getElementById(id).classList.remove('open');
  }

  function addActorRuleRow(containerId, key, values) {
    const container = document.getElementById(containerId);
    const row = document.createElement('div');
    row.className = 'actor-rule-row';
    row.innerHTML =
      "<input type='text' class='field-input actor-rule-key' placeholder='e.g. user_id' value='" + (key || '') + "' />" +
      "<input type='text' class='field-input actor-rule-values' placeholder='1, 2, 3' value='" + (values || '') + "' />" +
      "<button type='button' class='btn-icon actor-remove-btn'><i data-lucide='x' class='w-3 h-3'></i></button>";
    container.appendChild(row);
    row.querySelector('.actor-remove-btn').addEventListener('click', () => row.remove());
    lucide.createIcons();
  }

  function collectActorRules(containerId) {
    const actors = {};
    document.querySelectorAll('#' + containerId + ' .actor-rule-row').forEach(row => {
      const key = row.querySelector('.actor-rule-key').value.trim();
      const vals = row.querySelector('.actor-rule-values').value.trim();
      if (key && vals) {
        actors[key] = vals.split(',').map(v => v.trim()).filter(Boolean);
      }
    });
    return actors;
  }

  function actorSummary(actors) {
    if (!actors) return [];
    return Object.entries(actors).map(([k, v]) => k + ':' + v.join(','));
  }

  // --- API ---
  async function apiFetch(method, body) {
    const opts = { method, headers: { 'Accept': 'application/json', 'Content-Type': 'application/json' } };
    if (body) opts.body = JSON.stringify(body);
    return fetch('', opts);
  }

  async function loadFlags() {
    try {
      const res = await apiFetch('GET');
      if (res.ok) { featureFlags = await res.json(); }
      else { featureFlags = {}; }
    } catch (e) { featureFlags = {}; }
  }

  // --- Render ---
  function render() {
    const container = document.getElementById('featureFlagsContainer');
    const search = (document.getElementById('searchInput').value || '').toLowerCase().trim();
    const allEntries = Object.entries(featureFlags);
    const entries = search ? allEntries.filter(([name]) => name.toLowerCase().includes(search)) : allEntries;

    if (allEntries.length === 0) {
      container.innerHTML = "<div class='empty-state'>No feature flags yet. Create one to get started.</div>";
    } else if (entries.length === 0) {
      container.innerHTML = "<div class='empty-state'>No flags matching '" + escapeHtml(search) + "'</div>";
    } else {
      container.innerHTML = '';
      entries.forEach(([name, cfg]) => {
        const card = document.createElement('div');
        card.className = 'flag-card';

        const actors = actorSummary(cfg.actors);
        const actorPills = actors.map(a =>
          "<span class='pill pill-actor'><i data-lucide='user' style='width:10px;height:10px;'></i> " + escapeHtml(a) + "</span>"
        ).join('');

        const rolloutPct = cfg.rollout || 0;

        card.innerHTML =
          "<div class='flag-top'>" +
            "<div style='display:flex;align-items:center;gap:0.75rem;min-width:0;'>" +
              "<label class='toggle'>" +
                "<input type='checkbox' " + (cfg.enabled ? "checked" : "") + " data-flag='" + escapeHtml(name) + "' />" +
                "<div class='toggle-track'></div>" +
                "<div class='toggle-thumb'></div>" +
              "</label>" +
              "<span class='flag-name'>" + escapeHtml(name) + "</span>" +
              "<span class='pill " + (cfg.enabled ? "pill-enabled" : "pill-disabled") + "'>" + (cfg.enabled ? "On" : "Off") + "</span>" +
            "</div>" +
            "<div class='flag-actions'>" +
              "<button class='btn-icon edit-btn' data-flag='" + escapeHtml(name) + "' title='Edit'><i data-lucide='settings-2' class='w-4 h-4'></i></button>" +
              "<button class='btn-icon delete-btn' data-flag='" + escapeHtml(name) + "' title='Delete'><i data-lucide='trash-2' class='w-4 h-4'></i></button>" +
            "</div>" +
          "</div>" +
          "<div class='flag-meta'>" +
            "<span class='pill pill-rollout'><i data-lucide='percent' style='width:10px;height:10px;'></i> " + rolloutPct + "% rollout</span>" +
            "<div class='rollout-bar'><div class='rollout-fill' style='width:" + rolloutPct + "%'></div></div>" +
            actorPills +
          "</div>";

        container.appendChild(card);

        card.querySelector('input[type=checkbox]').addEventListener('change', async (e) => {
          const key = e.target.dataset.flag;
          try {
            const res = await apiFetch('PATCH', { key });
            if (res.ok) {
              const result = await res.json();
              featureFlags[key].enabled = result.enabled;
              render();
              toastr.success("'" + key + "' " + (result.enabled ? 'enabled' : 'disabled'));
            }
          } catch (err) { toastr.error('Failed to toggle flag'); }
        });

        card.querySelector('.edit-btn').addEventListener('click', () => openEditDialog(name));
        card.querySelector('.delete-btn').addEventListener('click', () => openDeleteDialog(name));
      });
    }

    // Stats (always based on all flags, not filtered)
    let total = allEntries.length;
    let active = allEntries.filter(([, c]) => c.enabled).length;
    document.getElementById('totalFlagsCount').textContent = total;
    document.getElementById('activeFlagsCount').textContent = active;
    document.getElementById('coveragePercentage').textContent = total > 0 ? Math.round((active / total) * 100) + '%' : '0%';

    lucide.createIcons();
  }

  function escapeHtml(s) {
    const d = document.createElement('div');
    d.textContent = s;
    return d.innerHTML;
  }

  // --- Create dialog ---
  function syncEnableAll(checkboxId, rangeRowId, sliderId, valueId) {
    var cb = document.getElementById(checkboxId);
    var row = document.getElementById(rangeRowId);
    var slider = document.getElementById(sliderId);
    var val = document.getElementById(valueId);
    if (cb.checked) {
      row.classList.add('disabled');
      slider.value = 100;
      val.textContent = '100%';
    } else {
      row.classList.remove('disabled');
    }
  }

  document.getElementById('createEnableAll').addEventListener('change', function() {
    syncEnableAll('createEnableAll', 'createRangeRow', 'createRollout', 'createRolloutValue');
  });

  document.getElementById('addFlagTrigger').addEventListener('click', () => {
    document.getElementById('createFlagName').value = '';
    document.getElementById('createEnableAll').checked = true;
    document.getElementById('createRollout').value = 100;
    document.getElementById('createRolloutValue').textContent = '100%';
    document.getElementById('createRangeRow').classList.add('disabled');
    document.getElementById('createActorRules').innerHTML = '';
    openDialog('createDialog');
  });
  document.getElementById('createCancelBtn').addEventListener('click', () => closeDialog('createDialog'));
  document.getElementById('createDialogBackdrop').addEventListener('click', () => closeDialog('createDialog'));
  document.getElementById('createAddActorBtn').addEventListener('click', () => addActorRuleRow('createActorRules'));
  document.getElementById('createRollout').addEventListener('input', (e) => {
    document.getElementById('createRolloutValue').textContent = e.target.value + '%';
    document.getElementById('createEnableAll').checked = (parseInt(e.target.value, 10) === 100);
  });

  document.getElementById('createSubmitBtn').addEventListener('click', async () => {
    const name = document.getElementById('createFlagName').value.trim();
    if (!name) { toastr.error('Flag name is required'); return; }
    if (featureFlags[name] !== undefined) { toastr.warning("'" + name + "' already exists"); return; }

    const rollout = document.getElementById('createEnableAll').checked ? 100 : parseInt(document.getElementById('createRollout').value, 10);
    const actors = collectActorRules('createActorRules');

    try {
      const res = await apiFetch('POST', { key: name, enabled: true, rollout, actors: Object.keys(actors).length ? actors : null });
      if (res.ok) {
        const result = await res.json();
        featureFlags[name] = { enabled: result.enabled, rollout: result.rollout, actors: result.actors };
        closeDialog('createDialog');
        render();
        toastr.success("'" + name + "' created");
      } else {
        const err = await res.json();
        toastr.error(err.error || 'Failed to create flag');
      }
    } catch (e) { toastr.error('Failed to create flag'); }
  });

  // --- Edit dialog ---
  document.getElementById('editEnableAll').addEventListener('change', function() {
    syncEnableAll('editEnableAll', 'editRangeRow', 'editRollout', 'editRolloutValue');
  });

  function openEditDialog(name) {
    const cfg = featureFlags[name];
    const rollout = cfg.rollout || 0;
    const isAll = rollout >= 100;

    document.getElementById('editFlagKey').value = name;
    document.getElementById('editDialogTitle').textContent = 'Edit: ' + name;
    document.getElementById('editEnableAll').checked = isAll;
    document.getElementById('editRollout').value = rollout;
    document.getElementById('editRolloutValue').textContent = rollout + '%';
    if (isAll) {
      document.getElementById('editRangeRow').classList.add('disabled');
    } else {
      document.getElementById('editRangeRow').classList.remove('disabled');
    }

    const container = document.getElementById('editActorRules');
    container.innerHTML = '';
    if (cfg.actors) {
      Object.entries(cfg.actors).forEach(([k, v]) => {
        addActorRuleRow('editActorRules', k, v.join(', '));
      });
    }
    openDialog('editDialog');
  }
  document.getElementById('editCancelBtn').addEventListener('click', () => closeDialog('editDialog'));
  document.getElementById('editDialogBackdrop').addEventListener('click', () => closeDialog('editDialog'));
  document.getElementById('editAddActorBtn').addEventListener('click', () => addActorRuleRow('editActorRules'));
  document.getElementById('editRollout').addEventListener('input', (e) => {
    document.getElementById('editRolloutValue').textContent = e.target.value + '%';
    document.getElementById('editEnableAll').checked = (parseInt(e.target.value, 10) === 100);
  });

  document.getElementById('editSubmitBtn').addEventListener('click', async () => {
    const name = document.getElementById('editFlagKey').value;
    const rollout = document.getElementById('editEnableAll').checked ? 100 : parseInt(document.getElementById('editRollout').value, 10);
    const actors = collectActorRules('editActorRules');

    try {
      const res = await apiFetch('POST', { key: name, rollout, actors: Object.keys(actors).length ? actors : null });
      if (res.ok) {
        const result = await res.json();
        featureFlags[name] = { enabled: result.enabled, rollout: result.rollout, actors: result.actors };
        closeDialog('editDialog');
        render();
        toastr.success("'" + name + "' updated");
      } else {
        const err = await res.json();
        toastr.error(err.error || 'Failed to update flag');
      }
    } catch (e) { toastr.error('Failed to update flag'); }
  });

  // --- Delete dialog ---
  function openDeleteDialog(name) {
    flagToDelete = name;
    document.getElementById('deleteFlagName').textContent = name;
    openDialog('deleteDialog');
  }
  document.getElementById('deleteCancelBtn').addEventListener('click', () => closeDialog('deleteDialog'));
  document.getElementById('deleteDialogBackdrop').addEventListener('click', () => closeDialog('deleteDialog'));
  document.getElementById('deleteConfirmBtn').addEventListener('click', async () => {
    try {
      const res = await apiFetch('DELETE', { key: flagToDelete });
      if (res.ok) {
        delete featureFlags[flagToDelete];
        closeDialog('deleteDialog');
        render();
        toastr.info("'" + flagToDelete + "' deleted");
      }
    } catch (e) { toastr.error('Failed to delete flag'); }
  });

  // --- Theme toggle ---
  function updateThemeIcons() {
    var current = document.documentElement.getAttribute('data-theme') || 'dark';
    document.getElementById('themeIconLight').style.display = current === 'dark' ? 'block' : 'none';
    document.getElementById('themeIconDark').style.display = current === 'light' ? 'block' : 'none';
  }
  document.getElementById('themeToggle').addEventListener('click', function() {
    var current = document.documentElement.getAttribute('data-theme') || 'dark';
    var next = current === 'dark' ? 'light' : 'dark';
    document.documentElement.setAttribute('data-theme', next);
    localStorage.setItem('flaggo-theme', next);
    updateThemeIcons();
  });

  // --- Search ---
  document.getElementById('searchInput').addEventListener('input', () => render());

  // --- Init ---
  document.addEventListener('DOMContentLoaded', async () => {
    updateThemeIcons();
    await loadFlags();
    render();
  });
</script>
</body>
</html>`
