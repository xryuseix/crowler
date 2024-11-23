# 🐦‍⬛ Crowler

**_A Generic Docker-based Web Crawler that can be used to crawl any website and extract data from it._**

## How to use

### 1. Edit the `config.yaml` file

[config.yaml](./app/config.yaml) is the configuration file for the crawler.

example:

```yaml
# 同時に取得するWorkerの設定
thread_max: 3

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
docker-compose up
```
