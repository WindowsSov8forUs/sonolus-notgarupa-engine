# NotGarupa 引擎设计

本文件记录引擎的设计。行为依据按优先级：原作 → GarupaEditor simulator 的扩展 → Sonolus 平台约束。2026-10-18 起，GarupaEditor 经过算法检查的 simulator（`simulator-check`，7fe388f）是主要参照：其原作路径视为原作行为，扩展路径定义扩展；与逆向结论冲突且该处未经算法检查时，保留逆向结论并在文中注明。资源名称与几何见 [RESOURCE_CONTRACT.md](RESOURCE_CONTRACT.md)。

## 1. 谱面输入与转换

谱面由 `sonolus-notgarupa-server` 的转换器 `levelconv` 离线转换为 LevelData（输入规则与算法见 `sonolus-notgarupa-server/docs/CONVERTER.md`，数据格式见 [LEVEL_DATA.md](LEVEL_DATA.md)）。转换器完成所有只依赖谱面的计算；引擎只做与运行时设置（音符速度、音符大小、屏幕比例）和输入相关的计算。

### 1.1 GarupaEditor 与原作的差异分类

| 差异 | 分类 | 引擎处理 |
|---|---|---|
| 音符位于非整数位置（`beat*48` 不取整） | 扩展 | 保留实数节拍，命中时刻 = `BeatToTime(beat)` |
| 小数 lane、场外 lane（< 0 或 > 6）、宽度（非方向键 width ≤ 7） | 扩展 | 按 lane 线性外推坐标绘制；判定区为各成员 lane 目标点的圆 |
| SV / TimingGroup（含 0 与负速度） | 扩展 | 每组分段常速滚动轴，见 §3 |
| 负 BPM（折叠时间） | 扩展 | 转换器计算有符号时间；Slide 节点时间逆序时按各节点独立运动 |
| Slide 中间/头部的 Flick、Directional（gesture 节点） | 扩展 | 按 GarupaEditor 的 gesture 规则判定（到达 VPL 后的手势，Perfect/Miss） |
| Hidden 头/尾（headless/tailless）、全 Hidden Slide | 扩展 | 头/尾不判定；全 Hidden 仅绘制，主体 alpha × 0.5 |
| Long 与 Slide 的区分 | 原作（a6fabc4） | 只由 `long` 标记决定，不再按形状规整 |
| 同位置排序 | 原作（d3d7ef8、bb580ce） | 转换器按 GarupaEditor 批记录（含 BPM 记录与 Slide 宽方向尾 / 手势的 Add 成员）建批，批内用原作 .NET IntroSort（不稳定）按 buttonType 排序（投影音符用跨度中心）；该顺序决定方向键成组、同时押线与实体次序 |
| Slide 族 A/B 交替 | 影响同时押线与方向键连线的同组判断 | 转换器按 slideOrdinal 奇偶记录并参与配对 |
| 已触摸未滑动的 Flick 7 帧后判 **Perfect** | **GarupaEditor 错误**（已确认）：原作 WaitState 超时调用的是 onMiss 槽，不是 forcePerfect | 按原作：7 帧后 **Miss** |
| 触摸 → 最近 lane，无纵向限制、无相邻 lane 回退 | **不是扩展** | 按原作：判定圆半径 √0.23136（宽屏 1.168 × lane 间距），最近 lane 无可判音符时依次试第二、第三近 lane |
| 分数归一化为 10 000 000 + N、无连击倍率 | **不是扩展**（产品计分） | 采用 Sonolus 百万分（§6），倍率保留原作结构；评级比例与 GarupaEditor 相同 |
| Long 头超时两次 Miss 被视为错误 | **不是扩展** | 按原作：两次 Miss，各半伤害 |
| 未触碰的宽方向键组没有 Miss 路径 | **不是扩展** | 按原作：整组一次 Miss |
| 第二根手指触碰已被占有的方向键组 → 完整性错误 | **不是扩展** | 按原作：各成员各自绑定手指，先完成者判定全组 |
| Auto 显示 FC/AP | **不是扩展** | 按原作：Auto 通关状态恒为 1，不显示 FC/AP |
| HUD 全部在最上层 | **不是扩展** | 按原作：Combo/判定文字/AddScore 在音符之下，分数/生命/暂停在最上 |
| HABAHIRO（宽音符、换道） | 原作 | Garupa JSON 直接给出宽度，无需 BMS 的宽音符合并；实现见 §9 |
| Bezier 预处理、fever 等其他指令 | 原作 BMS 专有 | Garupa JSON 不产生，不实现 |

## 2. 布局（原作 §C1）

Sonolus 屏幕坐标与原作 GameCamera 世界坐标相同（y 向上，高 2）。预处理中计算：

```
aspect = 屏宽/屏高；s = aspect / 9.578571319580078
har = clamp01((aspect − 1.778666615486145) / (2 − 1.778666615486145))
safe = aspect > 1.7786666 ? 0.9 · 1.778666615486145 / aspect : 1        （Android 全屏规则）
m = safe·(1 + 0.054·har)；P = safe·(1 + 0.07·har)
S = s·m；Yb = −3.45 + 0.25·har
G（判定线 y）= Yb·S；W（lane 间距）= 2.2·S；V（消失点 y）= (Yb + 8.870001)·S
Y0（出现行 y）= G + 0.95·(V − G)；Ly = 5.42·S
```

音符位置：`c = 1.1^((progress−1)·50)`，`k = 0.05 + 0.95·c`，`x = (lane − 3)·W·k`，`y = Y0 − c·(Y0 − G)`。
音符缩放：`vr = s·N·|Ly − y| / |Ly − G|`，`a = har·(minRatio[按钮数−1] − 0.996) + 0.996`，`local = vr·a + (1 − a)`，`world = local·P`；精灵尺寸 = 像素 / PPU × world。

UI 单位：`U·k`，`U = 2·aspect/1334`；锚定元素按安全区角定位。

## 3. 时间与滚动

- 命中时刻 `T = BeatToTime(beat)`（Sonolus `#BPM_CHANGE`）。原作的整数位置量化、子步与 `activateAdjust` 只造成亚帧差异，不复现。
- 到达时间 `arrival(speed)`：`speed ≤ 11.01` 时 `(speed−1)·(−0.5)+5.5`，否则 `(speed−11)/(−10)+0.5`；设置项 1.0–12.0，默认 5.0。
- 无 SV 时 `progress = 1 − (T − now)/arrival`。
- SV（扩展）：每个 TimingGroup 的速度为阶梯函数（`#Global` 事件按替换方式作用于所有组，同拍时组事件先、全局事件后，取最后者）。滚动轴 `C(t) = ∫ speed`，`progress = 1 − (C(T) − C(now))/arrival`。转换器为每组输出连续编号的 `TimingChange` 实体（起始时刻、速度、起点处 C、起点前 C 的前缀最大值），并为每个音符写入 `C(T)`。
- 组实体在 `updateSequential` 中移动游标并写 `C(now)`；音符在 `updateParallel` 只读。
- 生成时刻：首个满足 `progress ≥ 0` 的时刻，即 `Cmax(t) ≥ C(T) − arrival`；`Cmax` 单调，按变化点二分查找。可见判定：`progress ≥ 0` 且在屏幕内。判定始终按真实时间（SV 不影响判定窗口）。

## 4. 实体与 LevelData

Archetype、数据字段与计分链见 [LEVEL_DATA.md](LEVEL_DATA.md)。转换器在离线阶段完成所有只依赖谱面的计算，引擎只做与运行时设置和输入相关的计算。

## 5. 判定（原作 §B）

- 时间窗：`f = |Δt|·60`，银行家舍入为整数帧 `r`；`r < sweet+3` Perfect，`+6` Great，`+7` Good，`+8` Bad，否则不判定。sweet：头/单点 0，Long 尾 1。
- Slide 节点（位置判定）的时间等价：`τ = (t_vpl − t)·60`，`t_vpl = T + δ`，`δ = ceil(arrival·60)/60 − arrival`；`τ ≤ 3` Perfect，`≤ 6` Great，`≤ 7` Good，`≤ 8` Bad；晚侧恒为 Perfect，直到超时。按住经过中间节点的条件：手指在节点 lane 判定圆内且 `τ ≤ 1`。
- 自动 Miss：单点 `13/60 s`；Long 头超时两次 Miss（各半伤害）；Slide 头在 13/60 s 与“到下一可见节点一半”二者较早者；中间节点同理并在下一可见节点停止时失效（伤害 ÷5）；Flick 尾停在线上 7 帧后 Miss；其余尾 13/60 s。
- 超时起点：Long 头在越过判定线（progress > 1）后进入 Wait，只有 Wait 中的头会超时；Wait 中的头继续移动越过判定线（`NoteLong.WaitState` 调用 `NoteBase.Move`），超时后整条 Long 消失；按住时头停在线上，尾到线后停在线上直到判定。Slide 头在越过判定线且到达虚拟完美线（τ ≤ 0）后才进入等待。SV 下 progress 是视觉进度，所以速度为 0 停在线上的头要等轴恢复移动后才超时（GarupaEditor `stopSlideHeadAtJudgeLine`）。中间节点停在真实判定线，终点停在虚拟完美线，因此中间节点的“下一节点已停止”判断对终点使用虚拟完美线。
- 尾部手势块（ExecTouchMoved，调整值 B = 0）：按住时每次在普通中间节点上的 Moved 也会检查 Flick 族终点。刚经过最后一个中间节点时，终点手势在同一帧判定；仍有中间节点未判定时，终点的接触与原点照常推进，一旦 Flick 完成，就按松手处理，未判定的节点 Miss（中间节点 ÷5）。Directional 终点在此处重置原点，所以实际只有 Flick 终点会这样完成。
- Flick：按下时缓存等级；2D 位移 > 0.04 rate（0.0711 世界单位）完成；7 个 ExecuteFrame 内未完成 → Miss。方向键：x 方向 + 水平 > 0.01 rate；宽方向键另需 2D > count·0.01 rate。
- 松手：中间节点待判时该节点 Miss，Slide 存活并可由任何新按下接续；普通尾在“甜区”（半径² × 1.08）内松手按 τ 评级；flick 类尾未完成手势即 Miss。
- 一个触摸至多判定一个音符；已被占有的音符胜出时该触摸被消耗。
- 扩展（GarupaEditor）：
  - Slide 头与中间的手势节点按原作 Slide 尾族判定：成功类型 5/6/7（Flick / Directional / 宽 Directional），超时 Miss 为类型 8（中间 ÷5），松手 Miss 用节点自身的 5/6/7。按下时只绑定（原点 = 触点、grace = 0），并启动 flash 与保持音；之后与尾部相同：在节点判定范围内 grace 恢复为 8，离开则每帧递减；到达 VPL 前原点跟随手指；到达且 grace > 0 时完成手势判 Perfect，Slide 继续。手势节点的接触范围是 Directional 的锚 lane 或 Flick 的跨度。未判定的手势节点挡住接触点，节点停在原位保持手势精灵。
  - Hidden 头 Slide 的第一个可见节点不是按钮候选，而是在常规处理后对所有未结束的触摸检查：普通节点在判定圆内且 `τ ≤ 1` 判 Perfect，手势节点在 `τ ≤ 8` 时绑定；同一手指可同时登记多个 Slide。
  - 仅剩 Hidden 节点的中间节点按 13/60 s 超时（与 GarupaEditor 一致）。
- 同一帧内按下又松开的触摸：先按 Began 处理（点亮 lane、绑定 Long/Slide），再按 Ended 处理（lane 渐灭、松手判定）。
- 等待手势的 Flick / Directional 与保持中的 Long 头仍参加按钮仲裁；被选中时该触摸被消耗。投影（非整数 lane）音符与 lane 音符在同一轮竞争。
- 终点节点在 VPL 停止：终点 Flick 的 7 帧计数和终点相关超时从 VPL 起算。
- 技能音符（Good 以上）只播放 `SE_RHYTHM_TAP_SKILL`，代替普通判定音，并受同位置静音约束；Slide 判定不使用技能粒子。宽方向键的主体粒子在跨度的滑动起始端，手指粒子在被滑动的成员上。

### 5.1 Sonolus 判定映射

| 原作 | Sonolus Judgment | 说明 |
|---|---|---|
| Perfect | Perfect | |
| Great | Great | |
| Good | Good | |
| Bad | Good | 精度记录真实误差；连击、HUD 与引擎自绘生命按原作 |
| Miss | Miss | |

Sonolus 自身的生命值只能在 preprocess 中静态配置（`AddLifeScheduled` 与 Entity Life 仅 preprocess 可写）：各计分 archetype Miss −100（全长曲 −50），Long 头 Miss 0（头超时的两次 ½ 伤害合计记在尾上），Slide 中间节点 Miss ÷5（逐实体 Entity Life，Play 与 Watch 分别经 `play.Entity`、`watch.Entity` 写入各自的 Entity Life 块 4007 / 4006）。Bad 在 Sonolus 中与 Good 共用判定，无法扣 50；引擎自绘的生命槽完全按原作。

## 6. 计分与 HUD

- 计分采用 Sonolus 的百万分（arcade 分数），与 Next-RUSH 相同：HUD 显示的分数与 Sonolus 结果页的分数是同一个数。Sonolus 没有卡组，原作按卡组综合力计算的分数无法在 Sonolus 中有意义地复现，因此不再采用原作公式与综合力设置。
- 公式：每个判定得 `判定倍率 × (1 + 连续加成)`，总分归一化为 1 000 000（预处理时累加全 Perfect、理想连击的总权重，HUD 显示 `floor(已得 / 总权重 × 1 000 000)`）。倍率保留原作的结构：Perfect 1.1、Great 0.8、Good 0.5（Sonolus 的 Good 包含原作的 Bad，原作 Bad 为 0 分）；连续加成为每 50 个连续 Great 及以上 +0.01，至 550 连击封顶（+0.11），近似原作的连击修正 1.00…1.11（原作在 21、51、101…701 处分档）。同一组常量同时写入 Level Score 块（`SetBase`、`SetConsecutive(JudgmentGreat)`）与 HUD 计算（`arcadeWeight`），两者不会不一致。连续计数就是连击数。
- Auto Live 每个音符 Perfect，因此得满分 1 000 000。AddScore 弹出显示该判定使显示分数增加的值。
- 连击：Perfect/Great +1；Good/Bad/Miss 归零。连击数在 1 秒内无变化后隐藏（0655843）。
- 判定文字：同一帧的多个判定只显示原始结果最高的一个（相同取先到者）。
- AddScore：每个得分判定启动一个弹出，4 个根轮流使用；x 固定在 282，沿 y 上升（135 → 63 → 54 → 45，9 帧一段共 27 帧）；数字中心 = 前一中心 + 前一字符内宽 + 1。
- 分数数字：原作整数字号缩放，8 位字号 28 间距 22、9 位 27 / 21、10 位 23 / 18。
- 游戏结束（生命 0）：原作此时直接终止演奏。Sonolus 无法终止演奏（生命归零只判定关卡失败，玩家可以继续），这里不强行还原，做法与 Next-RUSH 相同：生命停在 0 不再回复，生命槽显示游戏结束状态，其余（判定、粒子、lane 光柱、连击、AddScore、音效）照常；关卡失败，不再播放通关演出。
- 评级刻度按原作比例 1 : 6 : 12 : 18 : 24（原作为 `300·(E+95)·(…)`）放到百万分制，SS 取 90 %（与 GarupaEditor 归一化规则的比例相同）：C 37 500、B 225 000、A 450 000、S 675 000、SS 900 000；分数条上限为 SS × 1.111111，即 1 000 000；达到 SS 后 HighRankEffect（Flash 乒乓 + 8 个 kira 按 3 s 循环）在动态面板裁剪内绘制。
- HUD 自绘，Sonolus 原生 UI 除菜单与（watch）进度条外全部隐藏；菜单按钮放在原作暂停按钮位置。
- 层级：舞台灯光/音箱 < lane 场 < Combo/判定文字/AddScore < lane 光柱 < 判定线 < 长条主体 < 同时押线 < 音符 < 方向箭头/flash < 分数/生命/Auto 标签 < FC/AP。粒子由平台绘制在所有精灵之上。
- 音符层内的次序用 Sonolus 多键 z（`sonolus.ZIndex`）逐键表达，避免把时刻与 lane 压进一个浮点数时丢失精度：z1 = 层，z2 = −时刻（较早的在上），z3 = −lane（较低 lane 在上），z4 = 同一音符内的子序（Flick 箭头在主体之上，多宽方向键越靠近滑动方向越在上）。按原作，Long 尾与 Slide 节点使用头的时刻；相连的 Directional 成员把 z3 改为“方向 × lane”，使更靠近滑动方向的成员在上（原作的 ±0.00001 深度微调优先于 lane 次序）。长条主体、同时押线与方向连线只按 −时刻 排序。
- 通关：原作在 BGM 结束时通关。关卡给出 `clearTime`（BGM 结束时刻，转换器 `-clear-time` 或 devserver `meta.json`）时在该时刻（且全部判定后）通关，否则“全部计分单位判定后 0.5 s”。状态 3 = 全 Perfect、2 = 连击未断、1 = 其他；Auto Live 按原作恒为 1（GarupaEditor 的 `naturalCompletionPresentation` 为 3，是其测试用的有意改动，不采用）。通关后隐藏场地、判定线、lane 光柱、连击、判定文字与 AddScore。状态 ≥ 2 时 FC/AP 文字：字母落下、假名展开、星（30 Hz 采样的路径、旋转、缩放与淡入）、半透明幽灵星、带抖动的闪电（AP 晚 1/30 s）；FULLCOMBO、CLEAR、CLEAR_VO 同时播放。通关 0.2 s 起按原作时间生成 11 个基础烟花粒子，FC/AP 在 0.2333 s 生成星形烟花。
- 通关退场：从通关时刻起，舞台 TRS 用 OutQuad 在 4 s 内从演奏姿态（缩放 0.92、y 0）回到入场起点（0.7、y 111）（Stage.outroAnimation）。灯光、音箱按该姿态绘制，背景四边形每帧经 RuntimeBackground 按同一姿态放置（背景以自身中心缩放，该中心在根下方 170·缩放）。背景画布覆盖 0.92 下的整个 bgImage 部件（1766.4 UI 单位宽），缩到 0.7 时仍能盖满屏幕。
- 保持循环音（SE_RHYTHM_TAP_LONG）同时最多 4 个，新的开始会停止最早的一个（原作 cue 上限，GarupaEditor originalCueLimitProfile），被停止的那一个松手时也不再播放释放音。Play 在关卡内存中记录 4 个句柄；Watch 与 #EFFECT_AUTO 由 Stage 的 preprocess 沿计分链按开始顺序排程（音符在 preprocess 中把持续结束时间记在开始的计分单位上）；旧转换器生成、没有计分链的关卡仍由音符直接排程，不受上限约束。单次音效无法中止，不受此限。
- 舞台：舞台根缩放 `0.92·(1 − 0.23·HAR)`（背景 UI 根，整数 activeHeight，无安全区系数）、上移 14·HAR；背景四边形按同一变换放置。12 盏灯按 7 项色表每 0.2 s 插值，浅色项每盏灯在预处理时随机选取一次。音箱按判定跨度含 lane 0 / 6 触发，使用原作 `speaker_normal` / `_exciting` / `_continuous` 动画片段（Root 缩放与跳动、半透明叠加层、两个发光的位置 / 缩放 / 透明度，60 Hz 采样）。可用“Stage lights”关闭。
- SS 特效：kira 按原作旋转（静态与动画），在面板裁剪两端 20 单位内淡出（Flash 无法做软裁剪）。
- 判定粒子：同一判定位置（半 lane）与同一预制体族的新判定先销毁旧粒子再生成（restartIfActive）。
- 曲线表（音箱、星、闪电）在预处理时写入关卡数据，运行时原地读取，避免每次动态下标复制整表。

## 7. 模式

- **Play**：见上。`#EFFECT_AUTO` 开启时所有判定音效改为在预处理中按谱面时刻调度（同类音效最小间隔 1 帧合并同时音符），保持循环音同样按谱面调度。
- **Watch**：非回放时为 Auto Live；回放时读取 `#JUDGMENT`、`#ACCURACY`，Good 依精度还原为 Good/Bad，判定时刻 = 命中时刻 + 精度。Stage 在 preprocess 中沿 `firstScored → scoreNext` 链逐个应用判定并把 HUD 快照写入各音符 shared memory，同时调度全部判定音效；运行时只按时间查找最近快照，因此任意跳转（`Skip`）都无需回滚状态。粒子与 lane 光柱在 Stage 的 updateSequential 中随游标前进触发（跳转帧不触发）。
- **Preview**：从左到右的列，每列 2 s，lane 竖直、时间向上；绘制 lane 线、拍线（每 4 拍加粗）、全部音符、Long/Slide 主体（跨列分段）、同时押线；列底打印时间，BPM 变化处打印 BPM。SV 不影响预览（按真实时间）。
- **Tutorial**：循环演示 Tap、Flick、Directional、Long、Slide 五段，带手势图标、说明文字、判定音效与粒子；上一个/下一个切换段落。

## 7.1 绘制与布局（GarupaEditor 算法检查后）

- 按住、停止或吸附在判定线上的精灵使用目标按钮缩放 `rate·N`（世界 ×P），不再用移动中的混合缩放（停止的 Slide 节点含终点，以及未按住、吸附在线上的 Slide 头）；主体前端在按住点用同一缩放；尚未出现的节点端点用虚拟缩放 `rate·N·|Ly/safe − Y0/P| / |(Ly − G)/safe|`。按住中的 Slide 接触点在按钮局部空间移动，世界坐标为 `(x, G)·P/M`。
- 方向键连线宽 = 局部缩放 × 0.75；判定线随音符大小缩放；音符大小设置 10%–200%（HABAHIRO 钳制到 80%–150%）。
- Long / Slide 子节点按头的时刻与 lane 排序；相连的方向键成员 z3 = −first + 2·方向·(lane − first)（first 为组最左 lane，即原作逐成员 ±0.00001 的链式深度）。lane 光柱 OffReserve 为 1 帧，渐灭可重启，7 条光柱同深度按 lane 排序。
- 中间 Flick 节点保持 `note_slide_among` 并带上滑图标。Slide 主体材质只看后续节点是否有 Hidden。
- Slide 的节点 Miss（头超时、中间节点超时、松手时待判的节点）按原作 `hideNote(root, current, true)` 去掉该节点及其之前的全部主体段（含从它出发的一段），flash 与 TapKeep 停止；之后到达的段照常绘制。Watch 按回放中的 Miss 时刻同样处理，并在第一次 Miss 时结束按住表现（回放无法得知重新按下）。
- 新设置：Long 主体亮度（10–100，默认 80，即透明度）与 Sudden（0–100）。Sudden 精确裁剪主体线、同时押线与方向连线，画出现条；音符精灵在中心进入遮罩时整体隐藏（Sonolus 无遮罩）；lane fit 需要贴图裁剪，未实现；lane 光柱不遮罩。

## 7.2 未实现项与原因

Sonolus 框架限制（引擎无法实现，sonolus-go 无法弥补）：

| 原作行为 | 框架原因 | 当前处理 |
|---|---|---|
| Sudden 遮罩部分遮住音符与 lane 光柱；Life 填充按比例裁剪 | 没有遮罩 / 模板 | 音符整体隐藏，光柱不遮，生命填充近似 |
| Sudden lane fit、SS Flash 软边裁剪 | `Draw` 只接受精灵与四边形，不能指定纹理子区域（UV） | 未实现；可由资源包预切条带后近似 |
| 生命 0 时终止演奏 | 引擎无法终止演奏或暂停 BGM | 与 Next-RUSH 相同：生命停在 0，关卡失败，演奏照常继续，不再通关 |
| 同一音效最多 4 个声部 | `Play` 返回 0，单次播放无法中止 | 单次音效不限制；保持循环音已实现上限（见 §6） |
| 粒子位于音符之下、加色混合 | 粒子没有 z，皮肤 / 粒子数据没有混合模式 | 粒子由平台绘制在最上层 |
| 同一帧中超时先于触摸处理 | 回调顺序固定为 updateSequential → touch | 按 Sonolus 顺序 |
| 按手指槽位区分的空按音效 | 触摸 ID 不是固定槽位 | 按触摸 ID 近似 |
| 跨 BPM 时的判定偏移 | 偏移在时间上施加 | 按时间偏移 |

此前可以实现但未实现的部分已经补上：舞台退场动画、保持循环音的声部上限、同时押线的尾按运行时速度过期、停止的终点与未按住的吸附头使用目标缩放、中间节点 Moved 上的终点手势（D15）、SV 下的超时起点（D17）。荧光棒不实现：普通 live 中它在舞台入场开始 1.5 s 后用 45 帧淡出，早于音乐开始，之后只会被 GarupaEditor 的演出命令（谱面中没有）重新唤起。

sonolus-go：上表没有一项受其阻碍。N8–N10 已在 v2.5.1 修复；N11（在实体视图上选取内嵌基类：崩溃或静默生成错误代码）已在 v2.5.2 修复并支持该写法，引擎的隐藏头 Slide 保持音登记已改回经辅助函数写入。

## 8. 验证

没有 Sonolus 客户端时，引擎用一个基于 sonolus-go 内部 EngineData 执行器的无头生命周期模拟器测试（不在仓库中，见总报告）：按 Sonolus 生命周期（生成队列、initialize、updateSequential、touch、updateParallel、despawn）执行 play/watch 回调，由脚本机器人按谱面产生触摸。已验证：

- 三张 Bestdori 官方谱面（459 / 935 / 2331 个计分单位）与自制扩展谱面（Hidden 头/尾、Hidden 节点、方向尾、Long 方向尾、宽方向键、SV、小数与宽音符）机器人全 Perfect，无运行时错误，所有计分实体正常回收。
- 时间偏移 ±0.06 s → Great（Slide 晚侧 Perfect）、+0.11 s → Bad/Good 与“过半”头超时、−0.125 s → 不判定（空击音）。
- 漏击：Single/Flick −100、Long 头超时 2×50、Slide 头 −100、中间节点各 −20、尾 −100，生命槽精确。
- Watch Auto 与回放：回放的最终分数、判定文字、生命与对应 play 运行完全一致；向后跳转后 HUD 正确重建。
- 选项组合（镜像、音符速度 1/12、大小 0.8/1.5、关闭特效/音效、自动音效）。
- Draw 调用以真实皮肤图集栅格化，检查了场地、音符、长条、HUD、SS 特效与 AP 演出的画面。

对照 GarupaEditor 算法检查后版本修改后（2026-10-18）重新验证：

- 三张谱面（官方 Expert 459 单位、官方 Special 834 单位含 79 个方向键、自制扩展谱面）与 HABAHIRO 测试谱面：play 与 watch 全 Perfect，无运行时错误；时间偏移 ±0.06 / +0.11 / −0.125 s 的判定分布与修改前逐项一致。
- 转换器：官方谱面的实体数与同时押线（按拍与 lane 比较）与修改前逐条一致；新增 Long 标记、HABAHIRO、方向键成组、IntroSort 与宽方向尾押线重定位的单元测试。
- 画面：HABAHIRO 宽音符、换道闪光与场地切换、普通谱面 HUD 均以真实图集栅格化检查。

## 9. HABAHIRO

- `Header #HABAHIRO` 的关卡强制使用 HABAHIRO 皮肤、粒子与背景条目（HABAHIRO 条目不在条目列表中；原作中方向键皮肤与方向键特效保留玩家设置，生成器为每种设置各生成一个 HABAHIRO 条目，devserver 按配置项 `habahiroDirectional` / `habahiroDirectionalEffect` 为每个请求选出 `notgarupa-habahiro-<方向键>`、`notgarupa-habahiro-<方向键><normal|light>`，背景为 `notgarupa-habahiro-normal`）；引擎用 `Exists()` 检查，资源包缺少 HABAHIRO 精灵 / 粒子时回退普通资源。
- 宽音符（宽度直接来自谱面）绘制在跨度中心（偶数宽在两 lane 之间），精灵为 `habahiro/note_<族>_<s>_…_<s+w−1>`，像素宽 308…1596；Slide 中间节点用 `note_slide_among_<w>`，上滑图标按宽度选 `note_flick_top` / `_2` / `_3`；Long 的按住闪光用真实 lane，Slide 的按住头与闪光用居中的 run。主体线使用 HABAHIRO 的 longNoteLine，网格半宽乘 widthRate（w = 2：1.05；w ≥ 3：1.05 + 0.03·HAR）。方向键不变。
- 判定与按钮不变；lane 光柱关闭；音符大小钳制 80%–150%。
- 换道（最晚的 `#LANE_CHANGE`）：到点后 1 s 的闪光（4 张贴图，smoothstep 关键帧），0.4167 s 时场地与判定线（43 px）换成 HABAHIRO 版本；无换道时全程为 skin00 场地。
- 判定粒子按音符宽度选 `_2…_7` 变体（Long 的 TapKeep 同样；Slide 的 TapKeep 用基础版），锚点为接收 lane（偶数宽在半 lane）。
