# NotGarupa 资源契约

本文件约定 `sonolus-notgarupa-engine` 与四个资源包生成器之间的资源名称、几何与使用方式。机器可读的完整清单由生成器每次运行时重新写出：

| 资源包 | 生成器 | 机器可读清单 |
|---|---|---|
| Skin | `sonolus-notgarupa-skin-generator` | `sonolus-notgarupa-skin-generator/contract.json` |
| Effect | `sonolus-notgarupa-effect-generator` | `sonolus-notgarupa-effect-generator/contract.json` |
| Particle | `sonolus-notgarupa-particle-generator` | `sonolus-notgarupa-particle-generator/contract.json` |
| Background | `sonolus-notgarupa-background-generator` | （无，见下文） |

引擎需要的资源名清单在本仓库的 `resources.json`：它列出引擎声明的全部 skin 精灵、effect clip 与 particle 效果名，每类分为 `ordinary`（普通条目与 HABAHIRO 条目都必须提供）与 `habahiro`（只有 HABAHIRO 条目另外提供）。清单由 `engine/resources_test.go` 从引擎源码中的 `sonolus.SkinSprite` / `EffectClip` / `ParticleEffect` 声明提取：HABAHIRO 专用名是皮肤的 `Haba*` 字段与粒子的宽度变体 `*W` 字段。修改声明后运行 `go test ./engine -run TestResourceList -update` 更新清单，测试保证两者一致。服务器启动时按清单检查实际加载的条目（devserver 的 `-resources`）。NotGarupa 不支持第三方资源，所以清单中的名字缺一不可。

每个生成器都是独立项目：各自有 `go.mod`，用到的公共代码（Bestdori / 官方 CDN 下载与缓存、Unity bundle 与 NGUI 精灵解析、图集打包、字体预渲染、音频解码、sonolus-pack-go 调用）放在各自的 `internal/` 下，可单独 `go build` 与运行。本文件由引擎维护（引擎是资源的使用方）。

## 素材来源

- **Bestdori**（`https://bestdori.com/assets/<region>/<bundle>_rip/<file>`；各服有独占素材，每个 bundle 按服务器优先级 jp → en → tw → cn → kr 取自第一个拥有它的服务器）：note / directional / field / judge 皮肤、打击音效（`sound/tapseskin/*`、`sound/common`）、背景（`ingameskin/bgskin/*`）。
- **原作官方资源 CDN**（`https://content.garupa.jp/Release/<版本>/Android/<bundle>`，资源版本固定）：判定粒子 bundle（`ingameskin/tapeffect/skin00`、`directionalflickskin00normal`、`habahiro`），粒子生成器直接解析其中的 ParticleSystem 与贴图。
- **原作 APK 内置资源**（不在 AssetBundle 中，无法联网获取）：HUD 图集与其 atlas 行、`sgm` 字体、lane tap 光柱、Sudden 线、分数槽高分段贴图、舞台灯光/音箱/荧光棒、DarkCover、开场信息、FC/AP 字样、`decide_1`、通关烟花组件树与贴图。随各生成器放在其 `assets/apk/` 下并编入程序。

生成器不依赖任何第三方工程。所有下载写入 `-cache`（默认 `.cache`），之后可用 `-offline` 完全离线重建。

## Skin

### 命名规则

| 形式 | 含义 |
|---|---|
| `<原作名>` | 原作精灵原名（`note_normal_3`、`judge_perfect`、`icon_number_big_7`、`gauge_base_score` …） |
| `<名>#L` `#M` `#R` | 只有左右 border 的 sliced 精灵切成三段 |
| `<名>#TL #T #TR #L #M #R #BL #B #BR` | 有上下 border 的 sliced 精灵切成九段 |
| `<名>@RRGGBB[AA]` | 原作运行时着色的预烘焙版本（含其切片 `<名>@hex#L` …） |
| `sgm_<字号>_<HEX>_<字符>` | `sgm`（TT-Shin Go）字体单字形，2× 超采样 |
| `sgmtext_<字号>_<HEX>_<id>` | 整段预渲染文字 |
| `habahiro/<原作名>` | HABAHIRO（多 lane 宽音符）专用精灵，只在 HABAHIRO 皮肤条目中，命名与几何见下文 HABAHIRO 节 |

重名处理：RhythmGameAdditive 的 `hp_meter` 为 `additive/hp_meter`；APK 资源文件名中的 hash 后缀已去掉（`AllPerfect_A`）；高分段贴图使用原名 `ss_kira`、`sss_star_long`、`ss_overlay`。

### 几何

每个精灵在 `contract.json` 中记录：

- `w`、`h`：贴图像素（字形为逻辑字号像素，贴图为其 `scale` 倍）。
- `pivot`：归一化轴心，左下为原点（Unity 约定）；来自 APK 图集 atlas 行（`render-catalog.json`）的 pivot 已由 `1 − pivotY` 换算到同一约定。
- `ppu`：Unity pixels-per-unit（NGUI 精灵为 0，以 UI 单位 = 像素绘制）。
- `border`：九宫格边距 `[左, 右, 上, 下]`。
- `trim` / `full`：透明边已裁掉时的裁切量 `[左, 上, 右, 下]` 与原始尺寸。`w`、`h`、`pivot` 已是裁切后的值，按 `w/ppu × h/ppu` 围绕 `pivot` 绘制即与原作位置一致。只有 `stage` 与 `clear` 两组（舞台灯光、音箱、通关文字等独立美术）会裁切；音符、判定、HUD、贴图类（`longNoteLine*`、`FlickNoteLine_*`、`bg_line_rhythm`）与 sliced 精灵都保持原尺寸。引擎按 `full` 矩形定位、再按 `trim` 内缩绘制（`stage_fx.go: drawTrimmed`）。
- `glyph`：字形的 `advance`、`ascent`、`descent`、`originX`（逻辑像素）。字形 `pivot` 位于基线上的笔位。

所有精灵变换为恒等变换，纹理插值开启。

### 分组

| 组 | 内容 |
|---|---|
| `note` | 45 个 note 精灵（`note_normal_*`、`note_normal_16_*`、`note_skill_*`、`note_flick_*`、`note_long_*`、`note_long_flash_*`、`note_slide_among`、`note_flick_top`、`simultaneous_line`），长条贴图 `longNoteLine`、`longNoteLine2` 及其 `@E6E6E6`（激活后的顶点 RGB 0.9 已烘焙，绘制时 alpha 取 0.8） |
| `directional` | `note_flick_l_*`、`note_flick_r_*`、`note_flick_top_l/_r`、背线贴图 `FlickNoteLine_l/_r` |
| `field` | `bg_line_rhythm`（1160×610 UI 单位，轴心底边居中，底边在判定线上）、`game_play_line`（PPU 69） |
| `judge` | `judge_perfect/great/good/bad/miss/auto/fast/slow`（NGUI，尺寸 = 精灵尺寸） |
| `hud` | RhythmGameUI（分数面板、分数条、生命条、暂停按钮、AddScore 数字、COMBO 字样…）、IconComboNumber（`icon_number_big_*`、`icon_number_big_AP_*`）、生命警告、SS 特效、`level_mark` |
| `lane` | `NoteLaneEffect_1..4`（PPU 69，轴心在完整 rect 底边）、`sudden_line` |
| `stage` | `Tex_miniShine`、`Tex_speaker`、`speaker_effect`、`Pyllium` 及其着色版本 |
| `clear` | `FullCombo_*`、`AllPerfect_*` 字样、星与雷电 |
| `text` | 预渲染字形与文字 |

### 着色与混合

Sonolus 精灵不能着色，也没有加法混合，因此：

- 原作运行时着色的精灵提供预烘焙版本：`hp_meter@6EFF69`（生命 > 20%）、`hp_meter@FF9999`（≤ 20%）、`additive/hp_meter@5DFF0064`（生命 > 1000 的第二条）、`label_round_white@FF3B72`（Auto 标签）、`ss_overlay@FFDDED`、8 个 `ss_kira@…`、`Tex_miniShine@0037FF` / `@D5FFFF`、`Pyllium@BFEEF0` / `@C8C8C8`、`longNoteLine*@E6E6E6`。
- 加法混合的素材（`additive/hp_meter`、`effect_health_caution_*`、`Tex_miniShine`、`speaker_effect`、`Pyllium`、`note_long_flash_*`）已就地转换：`alpha = max(rgb·a)`，`rgb = rgb·a / alpha`。在暗背景上与加法混合结果相同，在亮背景上会略压暗。hold flash 原作把 RGB 在 0.2→0.6→0.2 之间动画，绘制时把该值作为 alpha。

### 纹理尺寸

默认纹理 2048×4096（2 的幂）。Sonolus 文档只“建议”不超过 2048²；本皮肤内容裁边后约 4.9M px²，在不缩小原作素材的前提下放不进 2048²。`-max-texture` 可调整。可选内容：`-startup`（开场演出素材）、`-skill-adjust`（判定线技能叠加层）。HABAHIRO 条目见下文。

## Effect

片段名 = 原作 cue 名。默认格式 WAV：先按 MP3 的 LAME 标签去掉编码延迟与填充，解码长度与原作 HCA 采样数完全一致（例如 perfect 33001、SE_RHYTHM_TAP_LONG 23253），保证判定音与时间轴对齐。`-format mp3` 改为直接使用 Bestdori 的 MP3 字节（Sonolus 客户端是否接受 WAV 尚未在真机确认）。

| 片段 | 用途 |
|---|---|
| `perfect` `great` `good` | 标准判定音（Bad / Miss 原作无声） |
| `flick` | flick 系判定音（与标准判定音互斥） |
| `directional_fl` `directional_fl_2` `directional_fl_3` | 方向滑键（按宽度 1 / 2 / ≥3） |
| `game_button` | 空点（未命中音符的触摸） |
| `SE_RHYTHM_TAP_LONG` | 长按音效完整 cue |
| `SE_RHYTHM_TAP_LONG#loop` | 原作循环区间 0–22997 采样，`PlayLooped` 用 |
| `SE_RHYTHM_TAP_LONG#release` | 停止循环时播放的 0.3 s 线性淡出尾音（Sonolus 无法淡出循环） |
| `SE_RHYTHM_TAP_SKILL` | 技能音符（Great 以上，与判定音同时） |
| `SE_RHYTHM_CUTIN_SKILL` `SE_RHYTHM_CUTIN_AUDIENCE` | 技能发动 |
| `SE_RHYTHM_FULLCOMBO` `SE_RHYTHM_CLEAR` `SE_RHYTHM_CLEAR_VO` | 通关 |
| `SE_RHYTHM_GAYA` | 观众循环 |
| `decide_1` | 暂停/对话框按钮 |
| `miss` | 静音片段，仅供 `#MISS` |

标准名别名：`#PERFECT/#GREAT/#GOOD` → 对应 cue，`#*_ALTERNATIVE` → `flick`，`#HOLD/#HOLD_ALTERNATIVE` → `SE_RHYTHM_TAP_LONG#loop`，`#STAGE` → `game_button`，`#MISS*` → `miss`。

## Particle

原作 tap effect `skin00` 与方向滑键特效 `directionalflickskin00normal` 的每个 prefab 由 Unity ParticleSystem 序列化数据逐系统转换（发射、形状、速度限制/阻力、重力、角速度、大小/颜色随寿命变化、贴图序列帧、拉伸 billboard）。原作中处于 inactive 状态、`Play` 不会播放的 13 个子系统不转换。

### 坐标约定

- 特效坐标 = 原作粒子单位（一条 lane = 2.2 单位），原点 = 接收按钮的锚点（判定线上 lane 中心）。
- 引擎生成时传入以锚点为中心、半宽 `g`、半高 `g × quadScale.y` 的 quad，`g = rate · noteSize · safeRatio`（每粒子单位的世界单位）。
- `<prefab>#plane<度>` 是原作在 lane 平面上（绕 X 轴倾斜）的系统，引擎须以 `quadScale.y = cos(度)` 压扁 quad。与主特效同时生成、同样时长。
- 每个特效的时长写在 `contract.json`，生成时把该值作为 `duration` 传入。

| 特效 | 时长 (s) | 循环 | quadScale.y |
|---|---|---|---|
| `effect_tap`（空点） | 0.6 | 否 | 1 |
| `effect_tap_good` | 0.6 | 否 | 1 |
| `effect_tap_great` / `#plane50` | 0.7 | 否 | 1 / 0.6428 |
| `effect_tap_perfect` / `#plane46` | 0.9 | 否 | 1 / 0.6947 |
| `effect_tap_skill_good` | 0.6 | 否 | 1 |
| `effect_tap_skill_great` / `#plane46` | 0.7 | 否 | 1 / 0.6947 |
| `effect_tap_skill_perfect` / `#plane46` / `#plane50` | 0.9 | 否 | 1 / 0.6947 / 0.6428 |
| `effect_tap_swipe` / `#plane46` | 0.75 | 否 | 1 / 0.6947 |
| `effect_TapKeep` | 1 | 是 | 1 |
| `effect_TapKeep#plane48` | 6 | 是 | 0.6691 |
| `effect_tap_directional_flick_{l,r}` | 0.7 | 否 | 1 |
| `effect_tap_directional_flick_{l,r}_2` / `_3` | 1.1 | 否 | 1 |
| `effect_tap_directional_flick_{l,r}_finger` / `#plane46` | 0.35 | 否 | 1 / 0.6947 |

`effect_TapKeep#plane48` 的周期取 6 s，使旋转环（360°/s 与 300°/s）在循环边界处角度连续。循环特效按稳态生成，开始时已包含上一周期遗留的粒子。

### 与原作的差异（平台限制）

- Sonolus 粒子没有 z，统一绘制在精灵之上；原作粒子在音符之下（排序 1/5/50 < 70）。
- 原作为加法混合（线性色彩空间）；贴图已转换为“暗背景等效”的 alpha 混合版本，颜色转为显示空间。
- 粒子随机量中，Sonolus 只能线性表达的部分（方位角、起始大小、起始旋转、box 位置）保留真正的随机；其余（寿命、速度、颜色比例、贴图帧等）分层采样。
- 每个属性在每个时间段内线性插值，段数自适应（默认误差 0.05 粒子单位，最多 8 段）。

### 通关烟花（普通与 HABAHIRO 粒子条目都有）

由 `game-clear-profile.json` 转换。引擎在通关时刻 + `start` 生成一次，quad 以 `uiPoint(anchor)` 为中心，半宽半高固定 0.185 Sonolus 单位（不乘 UI 缩放），`quadScale` [1, 1]。

| 特效 | anchor (UI) | start (s) | 时长 (s) |
|---|---|---|---|
| `GameClearParticle/effect_par_firework_star_pink1` | (−457.545, −78.63) | 0.2 | 1.2 |
| `…_green1 (1)` | (−81.545, −28.63) | 0.533333 | 1.2 |
| `…_yellow1 (1)` | (−336.545, 93.37) | 0.8 | 1.2 |
| `…_blue1 (1)` | (345.455, −123.63) | 0.85 | 1.2 |
| `…_purple1 (1)` | (175.455, 99.37) | 0.983333 | 1.2 |
| `…_green1#effect_gameclear_firework_star_a` | (−98.545, −28.63) | 1.166667 | 0.5 |
| `…_green1#effect_gameclear_firework_star_b` | (121.05, −29.0) | 1.166667 | 1.2 |
| `…_pink1 (1)` | (−391.545, −78.63) | 1.45 | 1.2 |
| `…_blue1` | (346.455, −153.63) | 1.616667 | 1.2 |
| `…_yellow1` | (−284.545, 93.37) | 1.75 | 1.2 |
| `…_purple1` | (254.455, 99.37) | 1.85 | 1.2 |
| `FullComboAnimation/effect_par_firework_star`（状态 2） | (0, 45.873) | 0.233333 | 1.1 |
| `AllPerfectAnimation/effect_par_firework_star`（状态 3） | (0, 45.873) | 0.233333 | 1.1 |

### HABAHIRO 粒子条目

HABAHIRO 条目 `notgarupa-habahiro`（与普通条目同在一个包中）包含普通条目的全部特效，另加原作 `ingameskin/tapeffect/habahiro` 的宽度变体：`effect_tap_good` / `_great` / `_perfect`、`effect_tap_skill_good` / `_great` / `_perfect`、`effect_tap_swipe`、`effect_TapKeep` 各自的 `_2`…`_7`（`#planeNN` 部分与基础特效相同，时长相同）。引擎对宽度 w 的音符使用 `_<w>`，锚点为接收 lane（偶数宽在两 lane 之间）；Slide 的 TapKeep 始终用基础特效。

标准名：`#NOTE_*_TAP_*` → `effect_tap_perfect`，`#NOTE_*_ALTERNATIVE_*` → `effect_tap_swipe`，`#NOTE_*_HOLD_*` → `effect_TapKeep`，`#LANE_*` `#SLOT_*` `#JUDGE_LINE_*` → `effect_tap`（只复制主特效）。

## Background

`ingameskin/bgskin/<id>` 的 `liveBG_normal`（及 `liveBG_fever`）按原作 Standard live 舞台布局合成：bgImage 1920×1440 UI 单位、中心 y = −170、舞台根缩放 0.92，叠加 DarkCover（`CE/RevSubtraction`，在线性空间中减去灰 121 × linear(0.25)）。画布宽 1766.4 UI 单位（0.92 下的整个 bgImage 部件，宽于 1334 的屏幕）、高 1640 UI 单位，以屏幕中心对齐；引擎通过 RuntimeBackground 把它放在与 HUD 相同的 UI 单位中，并在通关退场动画（舞台缩放 0.92 → 0.7）中随舞台 TRS 移动，所以画布必须覆盖整个部件。DarkCover 纹理为均匀灰且覆盖所有可见位置，因此烘焙在整张画布上。舞台灯光、音箱、荧光棒有动画，由引擎用 skin 中的 `stage` 组精灵绘制。

高宽比 > 16:9 时原作会按 `1 − 0.23·HighAspectRatio` 缩小舞台根并上移 14·HighAspectRatio，静态背景无法随屏幕比例改变，因此在宽屏上背景比原作大。

## HABAHIRO

原作对 `#HABAHIRO` 谱面强制使用专用资源，不受玩家设置影响。HABAHIRO 是 NotGarupa 的固定组成部分：三个生成器每次都同时输出普通条目与 HABAHIRO 条目（同一个包），HABAHIRO 关卡通过 `useSkin` / `useParticle` / `useBackground` 强制选用（devserver 自动设置）。分成两个条目是因为一个关卡只能选一个条目，且 HABAHIRO 皮肤纹理为 4096×8192，普通关卡不必加载：

| 资源 | 条目 | 输出 |
|---|---|---|
| Skin | `notgarupa-skin00-skin00-habahiro`（普通全部精灵 + `habahiro/*`，纹理 4096×8192，全分辨率） | `sonolus-notgarupa-skin-generator/out/pack`，清单 `contract-habahiro.json` |
| Particle | `notgarupa-habahiro` | `sonolus-notgarupa-particle-generator/out/pack`，清单 `contract-habahiro.json` |
| Background | `notgarupa-habahiro-normal` / `-fever`（`ingameskin/bgskin/habahiro`） | `sonolus-notgarupa-background-generator/out/pack` |

宽音符精灵 `habahiro/note_<族>_<s>_<s+1>_…_<s+w−1>`（族：`note_normal`、`note_normal_16`、`note_skill`、`note_flick`、`note_long`、`note_long_flash`；每族 28 个 run，下标 `(w−1)(16−w)/2 + s`），像素宽 308 / 524 / 736 / 952 / 1164 / 1380 / 1596 × 120，PPU 100，pivot 中心；`habahiro/note_slide_among`、`_2`…`_7`；`habahiro/note_flick_top`、`_2`、`_3`（171 / 288 / 400 × 138）；`habahiro/longNoteLine*`、`habahiro/simultaneous_line`；换道后的场地 `habahiro/bg_line_rhythm` 与判定线 `habahiro/game_play_line`（1800×43）；换道闪光 `habahiro/bg_line_rhythm_flash`、`_flash2`、`habahiro/game_play_line_flash`、`_flash2`。几何、换道时间与闪光关键帧见 `sonolus-notgarupa-skin-generator/HABAHIRO.md`。
