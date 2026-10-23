# 工作歷程

## 2026-10-23 G0 啟動

- 目標：依使用者指示，參考銀色匕首之謎建立治理文件、private 儲存庫與初始 Issues；步驟完全比照銀色匕首之謎。
- 讀取：銀色匕首之謎 `AGENTS.md`、`CLAUDE.md`、`CONTEXT.md`、`README.md`、`.gitignore`、`WORKLOG.md`、`LICENSE`（含 G0 版 `dcd0115`）、
  `docs/spec/000-index.md`、`010-hd-theme.md`、`021-no-original-flash.md`、`docs/re/README.md` 與 Issue #1–#17；
  黑暗之后 `CONTEXT.md` 的 mode 13h 段落；幽靈騎士 `CONTEXT.md` 的 mode 與 HD 段落。
- 原始輸入：在 `sosb-tools:bookworm-r1`（`b9b6c7540eff`，沿用銀色匕首之謎；`--network none`、唯讀掛載）內計算兩個壓縮檔 SHA-256，
  以 `unzip -l` 列 zip 成員、`lsar -l -e cp950` 列 rar 成員，計算 `walkthrough/` 31 張掃描的 SHA-256（G1 記入 `docs/re/001`）。未解壓。
  成員清單暫存在本 session 的 scratchpad，不入 Git。
- 觀察：zip 是已展開的遊戲目錄，不是安裝器；附 `EGA.DRV`、`VGA.DRV`、`TGA.DRV`，前作都在 mode 0Dh 驗證，本作可能是 mode 13h。
  珍157 是 107 張 JPG，沒有 PDF。
- 使用者決定：repo `pools-of-darkness-cht`、短名 `pod`；G0 全做；授權上網取得英文手冊與手札；顯示模式 VGA 優先。
- dosgolem 候選以 `git for-each-ref` 與 `gh api` 回讀：銀色匕首之謎 `dosgolem/sosb` `5e9a961`；克萊恩英豪 `origin/dosgolem/cok` `29f7607`；
  黑暗之后 clone 的 `san5-mixer` `f79eab6`；dosgolem `main` `81c85be`。
- 變更：新增 `AGENTS.md`、`CLAUDE.md`、`.gitignore`、`LICENSE`（由銀色匕首之謎 G0 版 RRSAL-1.0 改填本作名稱與網址）、
  `CONTEXT.md`、`WORKLOG.md`、`README.md`。
- GitHub：`gh auth status` 為 `wicanr2`；同名 repo 不存在；建立 private repo 並推送，回讀 PRIVATE、Issues 啟用；
  建立 #1–#12（G0–G11）與 #13–#16（規格 018–021，對應銀色匕首之謎後來新增的同號需求），連結回填 AGENTS.md 第 2 節。
  銀色匕首之謎 #17（無敵位址驗證）併入 #11 的完成條件。
- 清理：容器皆 `--rm`，未寫入本目錄。
