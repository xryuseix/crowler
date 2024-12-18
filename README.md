# 🐦‍⬛ Crowler

**_A Generic Docker-based Web Crawler using chrome-devtools-protocol._**

<div align="center">
  <img src="./demo.gif" alt="Demo" />
</div>

## Output

Related files are also saved so that web pages can be restored.

```txt
out/example.com
├── index.html     // main page of the website(external links are replaced with local links)
├── index.old.html // original main page of the website
├── contents
│   ├── 0fae3c61-a154-4283-9216-fc2f63093f1b // js, css, and other files
│   ├── 2a854bf8-238a-4081-9ed4-98291b2682f6 // js, css, and other files
│   ├── ...
├── screenshot.png // screenshot of the website
└── url_table.json // external links and their corresponding local links
```

## How to use

### 1. Edit the `config.yaml` file

[config.yaml](./app/config.yaml) is the configuration file for the crawler.

example:

```yaml
# 同時に取得するWorkerの設定
# -1にするとruntime.NumCPU()の値が設定される
thread_max: -1

# 次のクロールまでの待ち時間の設定
wait_time: 0.5

# 開かないサイトの設定
## same-url...同じURLは二度と開かない
## same-domain...同じドメインのサイトは二度と開かない
## none...開かないサイトの設定をしない
duplicate: same-domain

# 取得する内容の設定
fetch_contents:
  html: true
  screenshot: true
  css_js_other: true

# 初期データの設定
seed_file: seed.txt
random_seed: false

# 出力先の設定
output_dir: out

# タイムアウトの設定
timeout:
  # ページの読み込みに対するタイムアウト
  navigate: 30
  # ページ内のコンテンツ取得に対するタイムアウト
  fetch: 5

# seedデータから幅優先探索で閲覧する深さの最大値
hops: 1
```

### 2. Edit the `seed.txt` file

[seed.txt](./app/seed.txt) is the seed file for the crawler.

example:

```txt
https://www.google.com/search?q=foo
https://search.yahoo.co.jp/search?p=bar
https://search.goo.ne.jp/web.jsp?MT=buz
```

### 3. Run

```bash
docker-compose up -d
```
