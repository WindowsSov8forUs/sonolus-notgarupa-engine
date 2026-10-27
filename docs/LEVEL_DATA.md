# Ex-Garupa 关卡数据约定

本文件约定 `sonolus-notgarupa-server` 的转换器（`levelconv`）写出、`sonolus-notgarupa-engine` 读取的 Sonolus LevelData：archetype 名、各实体的数据字段及其含义。引擎是关卡数据的使用方，本文件由引擎维护；转换器的算法见 `sonolus-notgarupa-server/docs/CONVERTER.md`。

## Archetype

| Archetype | 计分 | 作用 |
|---|---|---|
| `Initialization` | | 实体 0；`noteCount`、`firstScored`（按判定时刻排序的计分单位链表头）。布局、HUD 模型、Sonolus 计分与生命设置、UI 布局 |
| `Stage` | | 常驻：舞台灯光、音箱、lane 场、判定线、lane 光柱、HUD、SS 特效；watch 中预计算 HUD |
| `InputManager` | | play：每帧整理触摸、lane 解析与候选仲裁（原作 `GetPlayButton` + `GetMoveEndTimeNearestZeroNote`） |
| `#BPM_CHANGE`、`TimingGroup`、`TimingChange` | | 时间轴；TimingGroup 在 updateSequential 维护本组 `C(now)` |
| `TapNote` | ✔ | Single / Skill（`offbeat` 选 16 分音符色） |
| `FlickNote` | ✔ | Flick |
| `DirectionalNote` / `DirectionalMember` | ✔ / | 方向键组：组长计分，成员各自绑定手指；`leader`、`left`、`right`、`count`、`arrow` |
| `LongHead`、`LongTail` | ✔ ✔ | Long；头实体持有按住状态、主体、flash、循环音与 TapKeep |
| `SlideHead` / `SlideStart` | ✔ / | Slide 头（`SlideStart` = Hidden 头）：持有游标、手指、接触点、主体、flash、保持反馈，并判定所有节点 |
| `SlideNode` / `SlideHidden` | ✔ / | Slide 连接点：绘制自身精灵并保存自身结果；Hidden 节点只作几何顶点，永不生成 |
| `SimLine` | | 同时押线（转换器按原作配对算法生成） |

所有音符公共字段：`#BEAT, lane, spriteLane, width, spanStart, group, axisHit, skill, direction, scoreNext`（宽音符另有 `spriteStart`）。`Initialization` 另有 `habahiro`、`laneChangeBeat`（HABAHIRO 关卡，无换道为 −1），以及 `fullLength`（全长曲：原作 `DamageUtility.CalcDamage` 在乐曲 master 的 `musicDataType` 为 `full` 时把 Miss / Bad 伤害减半为 −50 / −25；由服务器按乐曲写入，引擎不再提供对应的玩家选项）和 `illustBackground`（关卡带自定义背景：按原作轻量模式的卡面背景显示，见下）。`SimLine` 带两端的绘制 lane `aLane`、`bLane`（宽方向 Slide 尾或其 Add 成员时为该成员的 lane），可选的 `aNoMargin`、`bNoMargin`（该端是方向键音符：音符皮肤的同时押线边距不作用于这一端，原作 NoteSyncLine 对 GameNoteType 10–19 不收缩），以及可选的 `speedMin`、`speedMax`（见下）。永不生成的实体（`SlideHidden`、关闭同时押线时的 `SimLine`）的生成顺序放在队尾，避免阻塞 Sonolus 生成队列。

## 计分链

`Initialization.firstScored` 指向第一个计分单位，每个计分单位的 `scoreNext` 指向下一个，按判定时刻排序。Watch 的 HUD 快照与保持音排程、Play 的 `#EFFECT_AUTO` 保持音排程沿这条链进行。旧转换器生成的关卡没有这条链：引擎仍能运行，但 Watch 的 HUD 不更新，保持音不受同时 4 个的上限约束。

## 同时押线的速度范围

同时押线的配对依赖音符速度（见 CONVERTER.md）。只在部分音符速度下存在的线带 `speedMin`、`speedMax`（音符速度选项的十分位，1.0 → 10），`SimLine` 只在当前速度落在其中时生成；缺省为所有速度。

## 自定义背景（`illustBackground`）

关卡的自定义背景是关卡自己的背景条目（`useBackground`，由服务器提供），按原作轻量模式（`InGameGraphicsMode.Light`）显示卡面插画的方式显示，`Initialization.illustBackground` 为真时引擎：

- 背景拉伸到原作背景控件（1920×1440 UI 单位）上，控件位于舞台 TRS 下的局部 y = −164（`BgManager.initLightMode` / `adjustPositionForLightMode`）；
- 舞台 TRS 保持开场的起始姿态（缩放 0.7、y 111）：原作 `InGameStageManager.introAnimation` 在轻量模式下不调用 `moveStageTRS`，结算时的 `outroAnimation` 移回的也是这一姿态；
- 不绘制舞台灯光、音箱与荧光棒（`LightBackgroundModule` 加载后调用 `removeStageEffects`：`DestroyPsylliumManager`、`RemoveStageBottomLight`、`RemoveSpeaker`；SD 小人原本就不绘制）；BgCover 保留；
- 背景亮度在开场时从 0.5 经 0.75 s 变为「背景亮度」选项（原作 `LightModeIllustBrightness × 0.01`，30–100、步长 10、默认 80）后保持；
- 背景条目不叠加舞台的 DarkCover（原作轻量模式隐藏 DarkCover 贴图）。

依据：日服 10.2.0 `libil2cpp.so` 的 ARM64（`RhythmGameStartData.get_BackgroundModuleType`、`LightBackgroundModule`、`BgManager`、`InGameStageManager`、`InGameStageControllerBase`、`LiveCoreSettingsLightModeBrightness`）。
