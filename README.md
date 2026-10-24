# sonolus-notgarupa-engine

用 [sonolus-go](https://github.com/WindowsSov8forUs/sonolus-go) 编写的 Sonolus 引擎，复刻 BanG Dream! ガールズバンドパーティ！ 的演奏画面与判定，并支持 GarupaEditor simulator 的谱面扩展（小数位置与 lane、宽音符、SV/TimingGroup、Hidden 头/尾/节点、Slide 中的 Flick/Directional 等）。

- 设计与原作对照：[docs/ENGINE_DESIGN.md](docs/ENGINE_DESIGN.md)
- 关卡数据约定（`sonolus-notgarupa-server` 的转换器写出、引擎读取）：[docs/LEVEL_DATA.md](docs/LEVEL_DATA.md)
- 资源名称约定（四个 generator 生成、引擎使用）：[docs/RESOURCE_CONTRACT.md](docs/RESOURCE_CONTRACT.md)

NotGarupa 分为六个独立项目：本引擎、`sonolus-notgarupa-server`（谱面转换器与本地服务器）以及 `sonolus-notgarupa-{skin,particle,effect,background}-generator` 四个资源生成器。

## 目录

| 路径 | 内容 |
|---|---|
| `engine/` | 引擎（`//go:build play`、`watch`、`preview`、`tutorial` 分模式；无 tag 的文件为共享资源与几何） |
| `docs/` | 设计文档与两份约定 |

## 构建

```sh
sonolus-go vet -m all ./engine          # 四个模式的静态检查
sonolus-go build -o notgarupa ./engine  # 输出到 dist/notgarupa/
```

`sonolus-go` 命令行与 `go.mod` 中的 sonolus-go 版本需保持一致（编译器在命令行中）。`engine/dev-level.json` 是 sonolus-go 开发服务器（`sonolus-go dev`）使用的开发关卡，由 `sonolus-notgarupa-server` 的 `cmd/levelconv` 生成。完整的本地试玩（关卡、资源包）用 `sonolus-notgarupa-server` 的 `cmd/devserver`，它默认读取本项目的 `dist/notgarupa`。

## 设置项

`#SPEED`、`#NOTE_SPEED`（1–12，原作速度曲线）、`#NOTE_SIZE`（10%–200%）、`#SIMLINE`、16 分音符配色、`#LANE_EFFECT`、FAST/SLOW、AP 连击配色、连击位置、`#MIRROR`、`#EFFECT`、`#EFFECT_AUTO`、舞台灯光、`#NOTE_EFFECT`、方向键特效（原作「左右フリックアイコンエフェクト」大 / 小）、Long 主体亮度（10–100）、Sudden（0–100）。
