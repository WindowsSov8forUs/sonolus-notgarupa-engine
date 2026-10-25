# 未生成的资源

资源生成器按原作 JP 10.1.4（230）的皮肤主数据生成全部条目：普通设置（音符、lane、打击特效、打击音效、方向键与其特效 normal / light）的排列组合，Collaboration / Limited 特殊组合（固定其组件，其余轴照常排列），以及谱面强制的 HABAHIRO 组合（规则见 `RESOURCE_CONTRACT.md`）。主数据快照在各生成器的 `internal/master/skins.json`，排除表在同目录的 `master.go`（`Excluded`）。

下列内容目前没有任何途径可以取得完整的原始资源，因此整体不生成：不做部分生成，也不用其他资源顶替缺失的部分。取得途径出现后，从排除表中删去对应项即可。

## 排除的主数据条目

| 条目 | 内容 | 原因 |
|---|---|---|
| Collaboration 3（`seasonSpecialId` 36） | `skinapril2019`：背景、打击特效、lane、音符、打击音效、判定、技能 cut-in | 打击特效与打击音效的 bundle 已不在官方 CDN 上（HTTP 403），Bestdori 的拆包不完整；原作同样把这一行视为不可用。后来复刻的 Limited 2 / Collaboration 22（`skin_april2019`）照常生成 |
| 活动背景 `skin01`（剧情活动） | `ingameskin/bgskin/skin01` | 官方 CDN 与 Bestdori 各服都没有这个 bundle |
| 活动背景 `skin04`（Live Try 活动） | `ingameskin/bgskin/skin04` | 同上 |
| 活动背景 `skin_teamlivefestival`（团队 Live 祭） | `ingameskin/bgskin/skin_teamlivefestival` | bundle 中没有 liveBG 贴图：只有团队 Live 祭的多人模式用各队的舞台物件拼出背景，单人 Live 不会显示它 |

## 不在范围内的服务器独占内容

生成范围是 JP 主数据。大陆服的特殊组合只存在于大陆服主数据中，并且缺少原始资源：

| 条目 | 组件 | 缺少的资源 |
|---|---|---|
| Limited 10001「小豆泥联动①」、10002「小豆泥联动②」，Collaboration 100014 | `skin_bilicat` / `skin_bilicat01`：背景、打击特效、lane、音符、判定、方向键 | 打击特效与方向键特效（`tapeffect/skin_bilicat*`、`directionalflickskin_bilicat*{normal,light}`）：JP CDN 没有（403），Bestdori 只有 JSON 导出与 PNG，没有可转换的 ParticleSystem 原始数据 |
| Collaboration 100001「彩虹节拍」 | `skin10001`：背景、lane、音符、打击音效、方向键 | 打击音效的 ACB（循环点与发音上限）：JP CDN 没有（403），Bestdori 不提供 ACB |

## 原作有、但引擎不使用的组件

这些不是资源缺失，而是 Ex-Garupa 引擎没有对应的功能，所以生成器不输出：

- 技能 cut-in 皮肤（只有已排除的 Collaboration 3 带有）。
- 舞台结构皮肤 `stageskin/normal|challenge|versus`：只用于 Live2D 演出路线。
- 判定线的技能叠加层 `game_play_line_skill_adjust_effect`，以及开场歌曲信息页的美术 `startup_line_star`、`startup_title_base`（引擎无法显示关卡封面与曲名，信息页不实现）。
- 菜单预览用的 `noteskin/*sample`、`bgskin/*preview`。
- 挑战活动与对邦活动背景的分层贴图 `liveBG_layer*`（`skin02`、`skin03`）：只用于挑战 Live、对邦 Live 模式；这两个背景按单人 Live 中显示的 liveBG 生成。
- 通关演出：FC / AP 字样（`FullCombo_*`、`AllPerfect_*`）、通关烟花粒子（`GameClearParticle/*`、`FullComboAnimation/*`、`AllPerfectAnimation/*`）、`SE_RHYTHM_FULLCOMBO`、`SE_RHYTHM_CLEAR`、`SE_RHYTHM_CLEAR_VO`。原作在 BGM 结束后播放，而 Sonolus 在 BGM 结束时结束关卡，见 `ENGINE_DESIGN.md`。
- 背景上的动态部分：开场 / 通关时的舞台动画、`skin_maid` 的飘雪粒子、循环帧动画。Sonolus 背景是静态图片，生成器按演奏中显示的状态合成。
