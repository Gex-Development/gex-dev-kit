#!/usr/bin/env python3

import argparse
import fcntl
import os
import pty
import pwd
import re
import select
import shutil
import struct
import subprocess
import sys
import termios
import time
from pathlib import Path

import pyte
from PIL import Image, ImageDraw, ImageFont

COLS = 100
ROWS = 30
CELL_W = 11
CELL_H = 20
GUIDE = Path(__file__).resolve().parent.parent
OUT = GUIDE / 'assets' / 'screens'
FONT_CANDIDATES = [
    '/System/Library/Fonts/Menlo.ttc',
    '/usr/share/fonts/truetype/dejavu/DejaVuSansMono.ttf',
]
REAL_HOME = pwd.getpwuid(os.getuid()).pw_dir
USERS_ROOT = '/' + 'Users' + '/'
PRIVATE_ROOT = '/' + 'private' + '/'
FOLDERS_ROOT = '/var/folders/'
TOKENS = [Path(REAL_HOME).name, USERS_ROOT, FOLDERS_ROOT, PRIVATE_ROOT, 'scratch' + 'pad']
TERMINAL_QUERY = re.compile(rb'\x1b\[[<=>?][\x20-\x3f]*[\x40-\x7e]')
BACKGROUND_QUERY = re.compile(rb'\x1b\]11;\?(?:\x07|\x1b\\)')
BACKGROUND_REPLY = b'\x1b]11;rgb:0b0b/1111/2020\x1b\\'
PALETTE = {
    'default': '#dbe5f3', 'black': '#152039', 'red': '#f87171',
    'green': '#64d2a0', 'brown': '#ffd166', 'blue': '#73b9ff',
    'magenta': '#c4a7e7', 'cyan': '#86dbeb', 'white': '#f5f8fc',
    'brightblack': '#8090a8', 'brightred': '#ff8b8b',
    'brightgreen': '#84e3aa', 'brightyellow': '#ffe19b',
    'brightblue': '#9ad3ff', 'brightmagenta': '#dbc4ff',
    'brightcyan': '#b2edf3', 'brightwhite': '#ffffff',
}
STATS = {'frames': 0}


def load_font(size):
    for candidate in FONT_CANDIDATES:
        if Path(candidate).is_file():
            return ImageFont.truetype(candidate, size)
    raise SystemExit('Nenhuma fonte monoespaçada encontrada.')


def sandbox_env(home, plain=False):
    env = {
        'HOME': str(home),
        'USERPROFILE': str(home),
        'CLAUDE_CONFIG_DIR': str(home / '.claude'),
        'CODEX_HOME': str(home / '.codex'),
        'XDG_CONFIG_HOME': str(home / '.config'),
        'XDG_DATA_HOME': str(home / '.local' / 'share'),
        'XDG_STATE_HOME': str(home / '.local' / 'state'),
        'XDG_CACHE_HOME': str(home / '.cache'),
        'PATH': '/usr/bin:/bin',
        'LANG': 'pt_BR.UTF-8',
    }
    if plain:
        env['TERM'] = 'dumb'
        env['NO_COLOR'] = '1'
    else:
        env['TERM'] = 'xterm-256color'
    return env


def count_tokens(text, home):
    return sum(text.count(token) for token in [*TOKENS, REAL_HOME, str(home)])


def redact(text, home):
    for token in [*TOKENS, REAL_HOME, str(home)]:
        text = text.replace(token, '[REDIGIDO]')
    return text


def check_privacy(screen, home):
    rows = []
    for y in range(ROWS):
        cells = [screen.buffer[y][x] for x in range(COLS)]
        text = ''.join(cell.data for cell in cells).rstrip()
        if count_tokens(text, home):
            raise PrivacyError(f'Portão de privacidade falhou na linha {y + 1}: {redact(text, home)}')
        rows.append((text, cells))
    return rows


class PrivacyError(RuntimeError):
    pass


def render(rows, name, title, command, footer):
    OUT.mkdir(parents=True, exist_ok=True)
    width = 80 + COLS * CELL_W
    height = 170 + ROWS * CELL_H + 40
    image = Image.new('RGB', (width, height), '#071529')
    draw = ImageDraw.Draw(image)
    strong = load_font(18)
    small = load_font(13)
    text_font = load_font(16)
    draw.rounded_rectangle((20, 16, width - 20, height - 16), radius=22, fill='#0d1729', outline='#29486c', width=2)
    for index, colour in enumerate(['#e98386', '#efcd79', '#73d2a2']):
        draw.ellipse((44 + index * 22, 36, 56 + index * 22, 48), fill=colour)
    draw.text((130, 30), title, font=strong, fill='#eaf3ff')
    draw.text((width - 190, 36), 'GEX · CLI', font=small, fill='#8bbde9')
    draw.line((22, 70, width - 22, 70), fill='#29486c', width=2)
    draw.text((44, 84), command, font=small, fill='#81c5f3')
    draw.text((44, height - 50), footer, font=small, fill='#879bb4')
    top = 120
    left = 40
    for y, (_, cells) in enumerate(rows):
        for x, cell in enumerate(cells):
            if cell.reverse and cell.data != ' ':
                draw.rectangle((left + x * CELL_W - 1, top + y * CELL_H, left + (x + 1) * CELL_W, top + (y + 1) * CELL_H), fill='#345475')
            if cell.data == ' ':
                continue
            colour = colour_of(cell.fg)
            draw.text((left + x * CELL_W, top + y * CELL_H), cell.data, font=text_font, fill=colour)
    target = OUT / f'{name}.png'
    image.save(target, optimize=True)
    return target


def colour_of(fg):
    if fg.startswith('#') and len(fg) == 7:
        return fg
    if len(fg) == 6 and all(c in '0123456789abcdefABCDEF' for c in fg):
        return '#' + fg
    return PALETTE.get(fg, PALETTE['default'])


def emit(name, screen, home, title, command):
    rows = check_privacy(screen, home)
    STATS['frames'] += 1
    text = [t for t, _ in rows if t.strip()]
    target = render(rows, name, title, command, 'EXECUÇÃO REAL  ·  HOME ISOLADO  ·  v0.3.0')
    print(f'{target.name}: sem vazamentos; primeira linha: {text[0] if text else ""}')


def feed(raw):
    screen = pyte.Screen(COLS, ROWS)
    pyte.ByteStream(screen).feed(TERMINAL_QUERY.sub(b'', raw))
    return screen


def answer_background_query(master, data):
    matches = BACKGROUND_QUERY.findall(data)
    for _ in matches:
        os.write(master, BACKGROUND_REPLY)
    return BACKGROUND_QUERY.sub(b'', data)


def read_for(master, seconds):
    chunks = []
    end = time.monotonic() + seconds
    while time.monotonic() < end:
        ready, _, _ = select.select([master], [], [], 0.1)
        if ready:
            try:
                data = os.read(master, 65536)
            except OSError:
                break
            if not data:
                break
            chunks.append(answer_background_query(master, data))
    return b''.join(chunks)


def spawn(binary, args, home):
    master, slave = pty.openpty()
    fcntl.ioctl(slave, termios.TIOCSWINSZ, struct.pack('HHHH', ROWS, COLS, 0, 0))
    child = subprocess.Popen([str(binary), *args], stdin=slave, stdout=slave, stderr=slave,
                             env=sandbox_env(home), cwd=str(home), start_new_session=True)
    os.close(slave)
    return master, child


def send(master, key, seconds):
    os.write(master, key)
    return read_for(master, seconds)


def stop(child, master):
    if child.poll() is None:
        child.kill()
        child.wait(timeout=3)
    os.close(master)


def capture_tui(binary, home):
    master, child = spawn(binary, [], home)
    try:
        raw = read_for(master, 1.6)
        emit('home-entrada', feed(raw), home, '01 · Ponto de partida', 'gex-dev-kit')
        raw += read_for(master, 2.5)
        emit('home-pronta', feed(raw), home, '02 · Estado do Core', 'gex-dev-kit')
        emit('home-ajuda', feed(send(master, b'?', 1.5)), home, '03 · Ajuda de teclado', 'gex-dev-kit · ?')
        send(master, b'\x1b', 1.0)
        emit('assistente', feed(send(master, b'\r', 1.5)), home, '04 · Instalação guiada', 'gex-dev-kit · Instalar')
        send(master, b'\r', 1.2)
        send(master, b' ', 1.0)
        for _ in range(5):
            send(master, b'\r', 1.2)
        send(master, b' ', 1.0)
        send(master, b'\r', 1.5)
        emit('previa', feed(send(master, b'n', 1.8)), home, '06 · Prévia antes de gravar', 'gex-dev-kit · Instalar')
        emit('confirmar', feed(send(master, b'\r', 1.5)), home, '07 · Confirmação', 'gex-dev-kit · Instalar')
        emit('gravando', feed(send(master, b's', 0.25)), home, '08 · Gravando', 'gex-dev-kit · Instalar')
        emit('concluido', feed(read_for(master, 3.0)), home, '09 · Instalação concluída', 'gex-dev-kit · Instalar')
    finally:
        stop(child, master)
    master, child = spawn(binary, [], home)
    try:
        read_for(master, 2.0)
        send(master, b'\x1b[B', 0.8)
        emit('catalogo', feed(send(master, b'\r', 2.0)), home, '10 · Catálogo', 'gex-dev-kit · Catálogo')
    finally:
        stop(child, master)


def run_plain(binary, args, home, name, title, command, codes, expect, keep='tail'):
    result = subprocess.run([str(binary), *args], cwd=str(home), env=sandbox_env(home, plain=True),
                            stdout=subprocess.PIPE, stderr=subprocess.STDOUT, timeout=120, check=False)
    text = result.stdout.decode('utf-8', 'replace')
    if result.returncode not in codes or expect.lower() not in text.lower():
        raise RuntimeError(f'{name} inesperado: saída {result.returncode}; trecho: {text[:300]!r}')
    lines = text.replace('\r\n', '\n').split('\n')
    kept = lines[:ROWS] if keep == 'head' else lines[-ROWS:]
    screen = feed(('\r\n'.join(kept)).encode('utf-8'))
    emit(name, screen, home, title, command)


def prepare_homes(workdir):
    if workdir.exists():
        shutil.rmtree(workdir)
    tui = workdir / 'tui'
    cli = workdir / 'cli'
    for home in (tui, cli):
        home.mkdir(parents=True)
    return tui, cli


def capture_cli(binary, home):
    run_plain(binary, ['skills'], home, 'cli-skills', '11 · Catálogo de skills',
              'gex-dev-kit skills', (0,), 'tdd', keep='head')
    run_plain(binary, ['install', '--skill', 'tdd', '--hook', 'gex-conventional-commits', '--skip-agent-update', '--yes'],
              home, 'cli-install', '12 · Prévia e gravação em modo CLI',
              'gex-dev-kit install --skill tdd --hook gex-conventional-commits --yes', (0,), 'Instalado. Histórico')
    run_plain(binary, ['doctor'], home, 'doctor', '13 · Diagnóstico',
              'gex-dev-kit doctor', (0, 1), 'Íntegro', keep='head')
    run_plain(binary, ['doctor', '--json'], home, 'cli-json', '14 · Saída JSON',
              'gex-dev-kit doctor --json', (0,), '"schema"', keep='head')
    run_plain(binary, ['rollback', '--yes'], home, 'rollback', '15 · Prévia e reversão',
              'gex-dev-kit rollback --yes', (0,), 'restaur', keep='tail')
    run_plain(binary, ['addons', 'caveman'], home, 'addons', '16 · Addon com comando oficial',
              'gex-dev-kit addons caveman', (0,), 'caveman', keep='head')
    run_plain(binary, ['ds'], home, 'ds', '17 · Vitrine do Design System',
              'gex-dev-kit ds', (0,), 'vitrine', keep='head')
    repo = home / 'repo-demo'
    repo.mkdir(exist_ok=True)
    subprocess.run(['/usr/bin/git', '-C', str(repo), 'init', '-q'], check=True, env=sandbox_env(home, plain=True))
    (repo / 'AGENTS.md').write_text('Comandos de verificação:\n\n- go vet ./...\n', encoding='utf-8')
    run_plain(binary, ['githooks', '--repo', str(repo), '--hook', 'gex-pre-commit', '--pre-commit-run', 'go vet ./...'],
              home, 'githooks-previa', '18 · Prévia dos git hooks',
              'gex-dev-kit githooks --repo <repositório> --hook gex-pre-commit', (0, 1), 'hook', keep='head')


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('binary', type=Path)
    parser.add_argument('--workdir', type=Path, required=True)
    args = parser.parse_args()
    binary = args.binary.expanduser().resolve()
    if not binary.is_file():
        raise SystemExit('Binário não encontrado.')
    tui, cli = prepare_homes(args.workdir.expanduser().resolve())
    failures = []
    for label, step in (('tui', lambda: capture_tui(binary, tui)), ('cli', lambda: capture_cli(binary, cli))):
        try:
            step()
        except PrivacyError:
            raise
        except Exception as error:
            failures.append(f'{label}: {error}')
    print(f'quadros: {STATS["frames"]}; vazamentos: 0')
    for failure in failures:
        print('falha:', failure)
    return 1 if failures else 0


if __name__ == '__main__':
    sys.exit(main())
