# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## Unreleased

## [v1.9.2](https://github.com/Duke1616/ecmdb/releases/tag/v1.9.2) - 2025-09-26

- [`1fca303`](https://github.com/Duke1616/ecmdb/commit/1fca303f3e5d806505c78038dc7d7cfdab3c1fb9) fix: 加载策略失败，不直接 panic 处理
- [`356a48f`](https://github.com/Duke1616/ecmdb/commit/356a48fa7e286a4bc2b9b573cf824c30c364531e) chore: 1.9.2 版本数据修复
- [`0636773`](https://github.com/Duke1616/ecmdb/commit/06367739fe098961889ec61d379cb9d1902f0bb7) chore: 鉴权 grpc 支持，resouce 传递
- [`71a1234`](https://github.com/Duke1616/ecmdb/commit/71a1234200369d19fbab944ce2802d1543813aa2) chore: logout 登出
- [`800d800`](https://github.com/Duke1616/ecmdb/commit/800d8003105f02dd1d8c5e9bdee4d26214ff0270) chore: session data 中存储 username
- [`a55ca5f`](https://github.com/Duke1616/ecmdb/commit/a55ca5f2b2fe1a1c737b563a979650c7926f9cbb) chore: 兼容 cookie 失败，通过 token 替代的情况
- [`1eb33e1`](https://github.com/Duke1616/ecmdb/commit/1eb33e15d2f1458fab1e576cf6c4b4712bf2c8fb) chore: ioc session 文件校验

## [v1.9.1](https://github.com/Duke1616/ecmdb/releases/tag/v1.9.1) - 2025-09-19

- [`360b985`](https://github.com/Duke1616/ecmdb/commit/360b985205eac094d538999d446908e43c7321a4) chore: 优化代码逻辑，提高健壮性
- [`27af294`](https://github.com/Duke1616/ecmdb/commit/27af29487a1cff6510ae5ed1bffcb272e95dabc4) fix: 修复资产数据加密逻辑，模型字段属性变更，资产同步变更，封装加密工具
- [`fb76249`](https://github.com/Duke1616/ecmdb/commit/fb76249c4afc0cc764cccdd764ae6c5d7f843b1b) fix: 资产部分接口增加数据加密解密, 重新封装 cryptox 方法
- [`e7f626f`](https://github.com/Duke1616/ecmdb/commit/e7f626f6332352187aa4702341bd34dcc4c116e1) chore: 删除打包二进制文件
- [`382a032`](https://github.com/Duke1616/ecmdb/commit/382a0322ae1a39392056b44273b5f795010e5259) fix: 修复资产数据加密字段未真实加密情况，对于历史存在数据提供修复方式
- [`92e5b45`](https://github.com/Duke1616/ecmdb/commit/92e5b45aff122877a4774c627623a55247730009) chore: 优化错误提示
- [`649938d`](https://github.com/Duke1616/ecmdb/commit/649938dc3d681a130e9ab60598fe6ab37f659ae9) chore: 修改获取进度图方式，代码更严谨

## [v1.9.0](https://github.com/Duke1616/ecmdb/releases/tag/v1.9.0) - 2025-09-15

- [`806715f`](https://github.com/Duke1616/ecmdb/commit/806715f73bae807d0b93faa9948d8734aacfbdcd) refactor: 排班系统，使用 username 存储，替换 id 统一管理
- [`fa9a169`](https://github.com/Duke1616/ecmdb/commit/fa9a169752e704ccc83bdeedf73d5fda1ba485cc) fix: 用户绑定角色，传递 id 错误
- [`6c8933b`](https://github.com/Duke1616/ecmdb/commit/6c8933b5a61ed3ea7d4ad450f492c8f55f3cfa8a) chore: 更新菜单
- [`8f72df4`](https://github.com/Duke1616/ecmdb/commit/8f72df422c010be2748234fe3d1e76e598a77364) chore: 菜单支持多平台
- [`1892627`](https://github.com/Duke1616/ecmdb/commit/18926272c506cf6510923601343afdd633ea66bd) feat: 提供更方便的初始化环境方式，提高用户体验度
- [`4a0236e`](https://github.com/Duke1616/ecmdb/commit/4a0236ea16f704cae4610f1ed4f832d5bfce62ad) chore: 调整 ctx 构造顺序
- [`a29b602`](https://github.com/Duke1616/ecmdb/commit/a29b6025788eb78f84dcb410e53e3ec23cebe946) fix: 修复 ctx 超时
- [`4b21895`](https://github.com/Duke1616/ecmdb/commit/4b21895b49c53722a2ac675dc4fd55c62735e9db) chore: 验证回调消息
- [`cb3e9e1`](https://github.com/Duke1616/ecmdb/commit/cb3e9e1946e0a125b12446153dc1821ea2578128) fix: 修复飞书查看流程进度无法正常响应
- [`0b58df5`](https://github.com/Duke1616/ecmdb/commit/0b58df5ac1fcac9f8def515ca26db6890e46140d) fix: 遗漏修改
- [`d549711`](https://github.com/Duke1616/ecmdb/commit/d5497112d6c5a2069a1720007f99f065632a8305) refactor: codebook 修改为存储用户名称取代原本的用户ID
- [`9f841ca`](https://github.com/Duke1616/ecmdb/commit/9f841caf282ce4550a18318b4df9922b018f5af1) chore: 提供根据平台获取菜单接口
- [`3bafdb3`](https://github.com/Duke1616/ecmdb/commit/3bafdb324880d88e7074781bb9e9599017e7e1df) fix: go mod 修改
- [`3d26128`](https://github.com/Duke1616/ecmdb/commit/3d261282b13af18228f6d666c8b4ac14059a936c) chore: session 过期时间修改为 30 天
- [`a8b3cd0`](https://github.com/Duke1616/ecmdb/commit/a8b3cd0c04ca01a815ff73e30e034417efea3da0) chore: 优化 middleware 传递
- [`7148c95`](https://github.com/Duke1616/ecmdb/commit/7148c95c4e242e588c412a1992e4720380227203) chore: 改为通过 cookie 进行认证
- [`565a904`](https://github.com/Duke1616/ecmdb/commit/565a904a64d48f6f1449e4b25776c6fbbf6f315b) chore: 升级 session 机制
- [`0c5bee3`](https://github.com/Duke1616/ecmdb/commit/0c5bee3159b68915e66bb2c10b13367a0d90b2e1) fix: 升级依赖为安全的版本
- [`6f497ac`](https://github.com/Duke1616/ecmdb/commit/6f497ac6af172a0417e4e4d71673cd4f7c2c4b79) refactor: 编译器提示优化
- [`2b0cb73`](https://github.com/Duke1616/ecmdb/commit/2b0cb73a5429bcdcca2a939264f899009b97a05c) refactor: 优化代码
- [`4d8c097`](https://github.com/Duke1616/ecmdb/commit/4d8c0973fd25f9118f67791f13bcfa534751ca6c) refactor: 重新组织 user 模块消息通知代码，新增告警转工单能力, 并通过工单模版进行消息发送
- [`e3ae6ab`](https://github.com/Duke1616/ecmdb/commit/e3ae6ab408bd2514c0db71e738ac7d205d53f1cb) chore: 为前端提供根据IDS获取模版列表接口
- [`bcaaa34`](https://github.com/Duke1616/ecmdb/commit/bcaaa34148ff3c20683068b525163052be636326) fix: 程序无法启动，多余使用字段未删除
- [`bf3fb2f`](https://github.com/Duke1616/ecmdb/commit/bf3fb2f21b7f590f83bb156578c9296c8cdbae5c) chore: 优化代码，Order 模块不再传递模版名称，通过 id 去查询
- [`f3cf101`](https://github.com/Duke1616/ecmdb/commit/f3cf101b52b920da641b565a169a0294f232dccd) feat: order 模块编写 grpc 集成测试
- [`815bbbc`](https://github.com/Duke1616/ecmdb/commit/815bbbcaf5361651481adeae015197c184017738) feat: order 模块编写 grpc 集成测试
- [`7acda34`](https://github.com/Duke1616/ecmdb/commit/7acda34df9e2b3468cf1694dd19a185056538954) chore: 配置文件同步
- [`810c69a`](https://github.com/Duke1616/ecmdb/commit/810c69a198405b4148345ccaa5622f5a8370e467) faet: 工单创建封装 GRPC 接口
- [`f19991d`](https://github.com/Duke1616/ecmdb/commit/f19991d44218e5a0042348369428a9b5e236c3ef) chore: 修复 model 模块集成测试
- [`0d80375`](https://github.com/Duke1616/ecmdb/commit/0d803756a315a0f819ed48b6fe74f208d20888d1) chore: 修复 model 模块集成测试
- [`8407eed`](https://github.com/Duke1616/ecmdb/commit/8407eede785423f976d39610a0f8bcc82e65b3e7) refactor: 重构流程引擎事件消息通知，让代码更优雅
- [`94c2390`](https://github.com/Duke1616/ecmdb/commit/94c23906e286cdb1725619a36864c1bf88bc89c4) chore: 添加任务自动通过标记位
- [`5114858`](https://github.com/Duke1616/ecmdb/commit/5114858672c25f500f9c2fcc36f971e172ffb37c) fix: 自动化任务 自动pass 改为以 Utime 时间，以防自动化延迟节点无法自动通过
- [`0abc5dd`](https://github.com/Duke1616/ecmdb/commit/0abc5dd28f249654973b95221ea8d1abafd52322) fix: 同步数据，确实字段补充
- [`39f99b3`](https://github.com/Duke1616/ecmdb/commit/39f99b3a6395979b4c7658d9338732bc2dc1a100) chore: 删除不必要的同步字段
- [`f9c3dc3`](https://github.com/Duke1616/ecmdb/commit/f9c3dc3e4c8d687d31429d92e76026df1afcc736) chore: 同步数据接口
- [`688f330`](https://github.com/Duke1616/ecmdb/commit/688f3303e1485ad80e66a6a1950c5bab48083e89) chore: 自动化任务调度节点，支持自动发现
- [`7a97d64`](https://github.com/Duke1616/ecmdb/commit/7a97d64fd22a98afdad9c7048c20f740e3c2512b) feat: runner 模块支持 by_ids, by_workflow_id 获取数据
- [`9db517b`](https://github.com/Duke1616/ecmdb/commit/9db517b41d7eec78a627fd1351b930764c51d5ad) feat: 自动发现 CRUD 接口
- [`cd35655`](https://github.com/Duke1616/ecmdb/commit/cd356555198c4cd65aaf9aa41008f4bd0fb70c2e) chore: 校验工单驳回、同意用户是否与任务一致，超级管理员跳过校验
- [`74ef395`](https://github.com/Duke1616/ecmdb/commit/74ef3952573cbb1641d129ebade67fe35a423d2b) chore: 自动化任务定时逻辑
- [`3fa6c53`](https://github.com/Duke1616/ecmdb/commit/3fa6c5327b80d505b2342371fee58c3d3c85d329) chore: 自动化任务定时逻辑
- [`32a52d5`](https://github.com/Duke1616/ecmdb/commit/32a52d5c12ff89ff9aeb23d7dfcc6324fe55f0f2) refactor: 新增接口，为前端提供更友好的模版提取字段信息
- [`dd3a5fc`](https://github.com/Duke1616/ecmdb/commit/dd3a5fcb04c86be98023324cf0a57039dbde4859) refactor: 并发处理任务
- [`88f36b3`](https://github.com/Duke1616/ecmdb/commit/88f36b364a3f855806d6ee51551febec42ed393b) chore: 重构解析 options 数据获取真实字段信息
- [`58d1bd2`](https://github.com/Duke1616/ecmdb/commit/58d1bd2fea5b99692f3919753e2ed034c8e1419a) chore: 栅格布局处理字段解析
- [`74d7400`](https://github.com/Duke1616/ecmdb/commit/74d7400402482a1d8c551518df70443bb906b542) fix: wantResult 阻止正常情况下的消息通知
- [`fccd3c4`](https://github.com/Duke1616/ecmdb/commit/fccd3c4fe39e6c124e4e104a6cd7d028fd2db7a7) refactor: 自动化任务获取, 封装接口
- [`7fb5cdc`](https://github.com/Duke1616/ecmdb/commit/7fb5cdca14b835d5b76dcd43cf01ec0b4cbb59fa) refactor: 优化流程事件部分，提升代码整洁度
- [`c36648c`](https://github.com/Duke1616/ecmdb/commit/c36648c9d063eff2c1b48328c69832aac698311a) fix: 多选匹配
- [`3470398`](https://github.com/Duke1616/ecmdb/commit/34703986f05d745ebee92a80ce09890c4b99a922) chore: 新增start阶段消息通知、自动化阶段发送方式多选
- [`1ecdfb7`](https://github.com/Duke1616/ecmdb/commit/1ecdfb7a1c3e8aa9666399778086d8d0194504ea) chore: 临时处理返回消息组合
- [`26a308d`](https://github.com/Duke1616/ecmdb/commit/26a308d1eea3d15c5d8d62a5fd5a510757d5d296) refactor: 修改 runner tags 返回结构

## [v1.8.0](https://github.com/Duke1616/ecmdb/releases/tag/v1.8.0) - 2025-03-18

- [`c1cff73`](https://github.com/Duke1616/ecmdb/commit/c1cff73d1bfcbdf3b03fc68f6dfedf368b1550e7) chore: 新增工单任务记录
- [`4c0912a`](https://github.com/Duke1616/ecmdb/commit/4c0912a50b02fe621724b4d9aeed1e7c16d4f839) fix: 多副本自动化任务恢复策略
- [`9c6fb77`](https://github.com/Duke1616/ecmdb/commit/9c6fb77780d20da0f4f51ff55a371b66f839b0f5) chore: 调整定时任务查询间隔时间
- [`d1b3d88`](https://github.com/Duke1616/ecmdb/commit/d1b3d88f08865d91787ad453e4b9de3a6af94176) feat: 完善定时自动化任务，重启恢复
- [`757c128`](https://github.com/Duke1616/ecmdb/commit/757c128f4ba8b65f36ded9aefacbd5d86c92a674) refactor: 优化自动化任务代码整洁度、添加runner执行器topic存储
- [`1910c96`](https://github.com/Duke1616/ecmdb/commit/1910c96d745959300219a1b5f4a5a6f5fb150af0) refactor: 重构自动化任务执行函数、提升代码整洁性
- [`63d2c98`](https://github.com/Duke1616/ecmdb/commit/63d2c98551e6bcab491179c1e471f72708b911b7) feat: 拦截前端工单任务添加定时字段，完成度 30%
- [`0981b85`](https://github.com/Duke1616/ecmdb/commit/0981b8507cd9e78ef4ad28069074f5d33f6f41ec) chore: 优化模版管理，用户组 format 展示
- [`7dd0a13`](https://github.com/Duke1616/ecmdb/commit/7dd0a131a5cb4f8d07e7312e98324c6c2f81ba0c) fix: 飞书发起撤销工单，状态变更
- [`aaab99b`](https://github.com/Duke1616/ecmdb/commit/aaab99b0cde299f8a0ad527848fd20ef50e0030a) chore: 细节处理
- [`6c74b09`](https://github.com/Duke1616/ecmdb/commit/6c74b09d404f94224f86a8242981d96fd0b0193b) chore: 完善功能代码，待测试验证
- [`a828854`](https://github.com/Duke1616/ecmdb/commit/a8288543e7565ddf7ec7b67b191dff94c4910de5) feat: 工作流针对不同的规则解析用户
- [`a2947eb`](https://github.com/Duke1616/ecmdb/commit/a2947eb58de89e6772ed26ff224e5592c3fd3c3e) chore: 升级 vuefinder 版本
- [`1293a6f`](https://github.com/Duke1616/ecmdb/commit/1293a6f9d80da0422c566c5570c0cdf1d3357620) chore: 升级 vuefinder 版本
- [`2a8fe89`](https://github.com/Duke1616/ecmdb/commit/2a8fe893490c78ba5625deb1f9778c9964ab529e) chore: 工单创建成功，发送消息给创建人
- [`1fc2739`](https://github.com/Duke1616/ecmdb/commit/1fc2739788b27d05d0a8acc846723fba9389c1db) fix: ssh session 连接业务逻辑错误
- [`65b5d07`](https://github.com/Duke1616/ecmdb/commit/65b5d07598d0267ec09fa63d30a3f6be00b24bda) chore: 优化
- [`37b7c9b`](https://github.com/Duke1616/ecmdb/commit/37b7c9b8261a7cfedaeda0063e5120cd6c6a0111) chore: 为前端提供表达式接口
- [`76b5bc4`](https://github.com/Duke1616/ecmdb/commit/76b5bc4d90b84408ab70608298c04cdda1e0492c) fix: 连续的两个condition网关, 导致处理失败
- [`2bb9ef6`](https://github.com/Duke1616/ecmdb/commit/2bb9ef6d0130945fe9eee6c42d61d21735e2aba0) titel 友好提示
- [`5a50b84`](https://github.com/Duke1616/ecmdb/commit/5a50b8437c0425e42ef05d10ff45625c8eefa5c6) chore: json 序列化
- [`64399db`](https://github.com/Duke1616/ecmdb/commit/64399db50f5f4b92735df5cc5df2a5b3efe8996a) chore: 传递系统用户信息到自动化任务
- [`b582ac1`](https://github.com/Duke1616/ecmdb/commit/b582ac101eee881c348a1160a31c82a199eddde6) chore: 传递系统用户信息到自动化任务
- [`d75538b`](https://github.com/Duke1616/ecmdb/commit/d75538bf43fe07e0a2f5d943871af348c95139e7) chore: 传递系统用户信息到自动化任务
- [`69bcccf`](https://github.com/Duke1616/ecmdb/commit/69bcccfe142b221594c1a15e030b4c12d039db84) chore: 自动化任务创建人信息传递
- [`4b351e3`](https://github.com/Duke1616/ecmdb/commit/4b351e39ddb17dd655fe8b8ecd37c190d51536d4) chore: 保存时机存在问题，end 节点处理异常
- [`ce35013`](https://github.com/Duke1616/ecmdb/commit/ce35013856ef5c92d0d55f46222d6ce89346f588) chore: 保存时机存在问题，end 节点处理异常
- [`5b27666`](https://github.com/Duke1616/ecmdb/commit/5b27666227c6640a5c077dbea54c212fc9d7ca7f) fix: 处理网关和任务节点错误
- [`ee0937e`](https://github.com/Duke1616/ecmdb/commit/ee0937e6a324f92da9e9dbb373f20c44f758cb29) chore: 中文字体展示
- [`524118a`](https://github.com/Duke1616/ecmdb/commit/524118a281094ec73ee54934eaf1373aed22d9c5) chore: 图片缩放错误
- [`d49b54b`](https://github.com/Duke1616/ecmdb/commit/d49b54b467551d3b1fc186ad332da15fd1de701f) chore: 部署测试效果
- [`0c36709`](https://github.com/Duke1616/ecmdb/commit/0c36709e09c8fa50d3b8fd9e5b46a6df34f654e1) chore: 部署测试效果
- [`2c70cb1`](https://github.com/Duke1616/ecmdb/commit/2c70cb1f1fa727657a2e7429e55c449d8cc6f888) chore: 图片清晰度
- [`1f0b1ad`](https://github.com/Duke1616/ecmdb/commit/1f0b1adf4907fd7732a4dfb54a40b7b950c163fe) chore: 图片清晰度
- [`c41497b`](https://github.com/Duke1616/ecmdb/commit/c41497bf96d9c0c0f59fe772ccbff62a9830b418) chore: 尝试修改配置，提高图片清晰度
- [`8c8dd5e`](https://github.com/Duke1616/ecmdb/commit/8c8dd5e9a962bfbcb102bd44b9c0565f762f98a2) chore: 修改镜像
- [`313824d`](https://github.com/Duke1616/ecmdb/commit/313824d1c3462e97fb066347ab95f816a9a360ad) chore: 测试chromedp
- [`617c2d7`](https://github.com/Duke1616/ecmdb/commit/617c2d7fe880e9eb8f7b78e1ed27d594be81ca34) chore: 替换基础镜像
- [`31e516b`](https://github.com/Duke1616/ecmdb/commit/31e516b21c4cdbb421357a854fe37906216a3c89) chore: 工单流程进度，完善基本的能力，支持流程图颜色标注，TODO 网关处理
- [`98d1897`](https://github.com/Duke1616/ecmdb/commit/98d18974fe5726b340ef142f2ca1573aa8ff90ae) chore: 工单流程进度查看
- [`cab01be`](https://github.com/Duke1616/ecmdb/commit/cab01be7a8a5d1d9b285833ea5eb16729c3895e0) chore: 支持飞书当前工单进度查看
- [`150c287`](https://github.com/Duke1616/ecmdb/commit/150c2876135027613478a12f7a3b26ec882b0789) chore: 支持飞书当前工单进度查看
- [`cf27db7`](https://github.com/Duke1616/ecmdb/commit/cf27db74b68e1752ca67244dfbf80b6f50498fac) chore: chromedp 模拟浏览器行为
- [`6334fd2`](https://github.com/Duke1616/ecmdb/commit/6334fd2cc97d84624dc74c2711c7a8ce536f9621) chore: chromedp 模拟浏览器行为
- [`4f1f7b9`](https://github.com/Duke1616/ecmdb/commit/4f1f7b9af2d3aba3f89822d4719a0ec99d1ca3f5) refactor: 重构消息通知，新增抄送功能
- [`5000865`](https://github.com/Duke1616/ecmdb/commit/500086503a490be8271faca5b71ea081c1bc6c7c) chore: 优化消息通知模块
- [`bae0d8e`](https://github.com/Duke1616/ecmdb/commit/bae0d8e2484e8b299fa33951714052c821d68ee2) chore: 消息通知
- [`c95ca3c`](https://github.com/Duke1616/ecmdb/commit/c95ca3cf574b8796c83e5e23b51997e4335e6ada) chore: 过滤消息通知，排除没有必要的信息
- [`753cbb2`](https://github.com/Duke1616/ecmdb/commit/753cbb2ff9f2283596631801b81521d0c01335e3) chore: 日常工作
- [`7bba5c8`](https://github.com/Duke1616/ecmdb/commit/7bba5c822a974b0c61f6ca5b417d5e07a04e174f) chore: 对接 web sftp
- [`d3a1a69`](https://github.com/Duke1616/ecmdb/commit/d3a1a69f847f45f804ce38ef1674bedc3a4f181f) chore: 优化 ssh 连接，设置超时时间
- [`6e71bf5`](https://github.com/Duke1616/ecmdb/commit/6e71bf569d182fb03f8e624b87514819b38efc0e) refactor: 抽象 session 进行 ssh client 存储
- [`685438b`](https://github.com/Duke1616/ecmdb/commit/685438b7d5223dde1fcf380796f435edc672bc88) chore: 优化错误处理
- [`a01f707`](https://github.com/Duke1616/ecmdb/commit/a01f7076d5fec06a8cfd839a340990a23f32ab60) chore: 优化代码，报错处理、业务逻辑等
- [`e850242`](https://github.com/Duke1616/ecmdb/commit/e8502427f9cc0ae96a7c8cc6900baf35fab9e175) chore: 通过参数传递
- [`01169dc`](https://github.com/Duke1616/ecmdb/commit/01169dc072a9682fb26aebbc19f788198ba1e240) feat: CMDB 中数据集成登录认证
- [`fc2df71`](https://github.com/Duke1616/ecmdb/commit/fc2df717b88ceccf9036356db0564be1d5ed5fa3) feat: web terminal 后端 demo 版初步实现，封装gunc和sshx
- [`53c5dce`](https://github.com/Duke1616/ecmdb/commit/53c5dceea5900e8148cde34c0d31ca469bf949a3) chore: 调整部署流程，并验证可靠性
- [`8824c3e`](https://github.com/Duke1616/ecmdb/commit/8824c3e66f9e880e9067b6daef292b1fe11e7a0f) chore: 替换 mongo 部署镜像
- [`eb72e57`](https://github.com/Duke1616/ecmdb/commit/eb72e57acf0df47e3ceeac03689f8e68e703c069) chore: 为前端提供展示查询接口
- [`da33f93`](https://github.com/Duke1616/ecmdb/commit/da33f937c1ef1dc9d894ab5f3a59649ea823d2bb) feat: CMDB 支持文件类型， 支持单独修改指定属性字段数据
- [`b17152d`](https://github.com/Duke1616/ecmdb/commit/b17152d2c6fe6cdcb20e04093d26a0079bd1dcf2) chore: 延长 Minio 签名过期时间
- [`f278d36`](https://github.com/Duke1616/ecmdb/commit/f278d36714ce31a6d66eb42ab2590bef57fccc17) chore: 强制浏览器进行下载
- [`3156f9b`](https://github.com/Duke1616/ecmdb/commit/3156f9b34472c6c676b5e7350740300d06127a13) chore: 文件删除功能
- [`1d9ad88`](https://github.com/Duke1616/ecmdb/commit/1d9ad8849d341482f11d4605b43d8deadb9ea6ff) chore: tools 目录结构
- [`0fe28ad`](https://github.com/Duke1616/ecmdb/commit/0fe28add86b111b1e04cb0cef7eb44e0774dfd0c) chore: minio 生成签名
- [`45aa2ae`](https://github.com/Duke1616/ecmdb/commit/45aa2ae90eef4719cd55e399c17ad160c65427e7) chore: 添加minio依赖注入
- [`e1df601`](https://github.com/Duke1616/ecmdb/commit/e1df6013cbfed5b81833d65714974ca50188e8d4) chore: 消息格式错误
- [`50a6307`](https://github.com/Duke1616/ecmdb/commit/50a63074373b9b4ea990aced83e16007591f0333) chore: 自动化任务失败，发送消息通知 codebook 负责人，增加手动结束任务接口
- [`b070763`](https://github.com/Duke1616/ecmdb/commit/b07076352693b3bd2ea11d6bc68a9e6d55d5364f) chore: 优化文件上床, 处理大文件
- [`85042e7`](https://github.com/Duke1616/ecmdb/commit/85042e75c6a3eee849e7a57817e797e5cf19bd37) chore: 实时写入数据

## [v1.7.0](https://github.com/Duke1616/ecmdb/releases/tag/v1.7.0) - 2024-11-22

- [`c7177ba`](https://github.com/Duke1616/ecmdb/commit/c7177ba505c8f277f220a2e4caf8b2ce4f8d4ba8) chore: 排班系统功能性开发完成，替换最新版本权限菜单, 进行初步发版
- [`547e776`](https://github.com/Duke1616/ecmdb/commit/547e776aaccc1129ace77f258dc63604f3702e44) chore: 新增查询当前排班接口，兼容以小时为单位 resulttime 赋值
- [`0215458`](https://github.com/Duke1616/ecmdb/commit/021545856894edad7bc29533f837b04436bb18c7) chore: web 接口路由变更
- [`4e401a9`](https://github.com/Duke1616/ecmdb/commit/4e401a9f2d5e03e8af129fe972dd6ef64dbfea51) chore: 优化 rrule 计算排班表代码
- [`09c6494`](https://github.com/Duke1616/ecmdb/commit/09c6494ee53666705383d5a9056d9fbae62085bb) chore: 规则截止时间配置
- [`22b20e8`](https://github.com/Duke1616/ecmdb/commit/22b20e85d5f483abf295ccd5c620d85dd65644c6) feat: 排班系统完善临时调班功能
- [`67219b9`](https://github.com/Duke1616/ecmdb/commit/67219b9b879213b6f545fd4d5e9169f4b968b8fe) chore: 规避月末 etime 传输不够
- [`c1fcab9`](https://github.com/Duke1616/ecmdb/commit/c1fcab947b3ad8240256651ad5e6b5798660b049) chore: rrule 当前下期排班信息
- [`1de0521`](https://github.com/Duke1616/ecmdb/commit/1de052197ce603b674e7d1abb66324431d43d0d4) chore: 后端 rrule 排班规则, 初步版本
- [`02134aa`](https://github.com/Duke1616/ecmdb/commit/02134aa2c50146fe97bea4c31b2b8dde5dddee55) chore: 排班支持修改删除规则
- [`e9de143`](https://github.com/Duke1616/ecmdb/commit/e9de1430d9ca4491f0f49d96c0dd393527dd790a) chore: 排班系统提供接口支持
- [`a033619`](https://github.com/Duke1616/ecmdb/commit/a03361965de2090f3dbdedfa965f902780c28c75) feat: 排班系统、完成部分接口
- [`c0dcf20`](https://github.com/Duke1616/ecmdb/commit/c0dcf20ac8c2bb17db8707d508172c5253c2c3fe) fix: 修复企业微信token失效, 导致无法正常处理业务逻辑
- [`e2902d6`](https://github.com/Duke1616/ecmdb/commit/e2902d6b4d84d78b695cb560c58a93844b467172) chore: 事件报错日志提示
- [`16b0c9a`](https://github.com/Duke1616/ecmdb/commit/16b0c9a83fad5496811ffb4c2ab694e5be6fe143) chore: 消息通知 title 信息
- [`7b1dca9`](https://github.com/Duke1616/ecmdb/commit/7b1dca9ca48c85b630fb011a8131d48b8b72c715) chore: 提供飞书回调审批后续消息处理
- [`600b02f`](https://github.com/Duke1616/ecmdb/commit/600b02f62f9c4acdfbeb63a2f8dc6a7ecbbcc605) chore: 回调消息，新增工单ID
- [`7f5f0a5`](https://github.com/Duke1616/ecmdb/commit/7f5f0a5f1b83cbbf53d0739197967e2ce4d28d51) fix: 数据 map key值存储和缓存不一致，导致批量删除
- [`805d67a`](https://github.com/Duke1616/ecmdb/commit/805d67a04e40d3041af9bc21ece0a0b8421fff1c) chore: Prefix 区分索引
- [`c328797`](https://github.com/Duke1616/ecmdb/commit/c3287973ed05450c17cd927922d090bf5ab454f9) chore: 缓存数据一致，当 LDAP 删除用户了，重新同步缓存也要进行删除
- [`b9fe762`](https://github.com/Duke1616/ecmdb/commit/b9fe762e45952b0c6f34b8b59478588550596b79) feat: 使用 redisearch 支持全文检索， 增强查询能力
- [`9b137f9`](https://github.com/Duke1616/ecmdb/commit/9b137f9790d2de57b362a6f8a02a06e3c17c9637) chore: ioc redisearch
- [`8555b60`](https://github.com/Duke1616/ecmdb/commit/8555b60e42844c88dfc45fde9b511a69beaadd6d) Merge branch 'main' of https://github.com/Duke1616/ecmdb
- [`027ab6e`](https://github.com/Duke1616/ecmdb/commit/027ab6e6f15d928a7bec0aac9ac87e1c46e05ba2) feat: 同步 LDAP 用户，导入到本系统中
- [`19a549c`](https://github.com/Duke1616/ecmdb/commit/19a549c2fefa30f4aac6221d233da3a3e6a82954) chore: 通过前端 windows.location 获取当前 url 地址
- [`c196bbc`](https://github.com/Duke1616/ecmdb/commit/c196bbc174c86f62a2a122c29fec79ec8f4e2d2d) chore: 接口地址
- [`1703f9d`](https://github.com/Duke1616/ecmdb/commit/1703f9dae344ffc641b28ef689b45019079a85a8) chore: downlaod 接口
- [`3e2ea16`](https://github.com/Duke1616/ecmdb/commit/3e2ea1636d6f77ad9719a5d5a7d695d20b2f9ce9) chore: 封装工具类方法，文件上传
- [`09e37af`](https://github.com/Duke1616/ecmdb/commit/09e37af456049f51db94526c3d8c3ac536e5732b) chore: 流程录入模版名称变量
- [`0f63e78`](https://github.com/Duke1616/ecmdb/commit/0f63e78446e32dd6023d2b64d2a9dbdde554bc03) chore: 封装 automation 消息通知，写烂了！！！ 后续重构吧
- [`44a8c57`](https://github.com/Duke1616/ecmdb/commit/44a8c57007db9d279e06289ec9575f55944a98fc) refactor: 重构消息通知, 拆解不同流程节点，处理通知类型不同
- [`a97201b`](https://github.com/Duke1616/ecmdb/commit/a97201b4922cc11ef45e4eaa65c1f18f3663a23c) chore: 工单系统模版功能完善
- [`12a1e21`](https://github.com/Duke1616/ecmdb/commit/12a1e212533d71d498b0a6cf2b14aecebaead73d) chore: 模版支持修改 Desc 描述信息
- [`56934a6`](https://github.com/Duke1616/ecmdb/commit/56934a6011ba16e8b5bb63b95777bcdc1962504d) chore: 工单创建，当用户查询为空异常处理
- [`91d7b2a`](https://github.com/Duke1616/ecmdb/commit/91d7b2af560170aec29249bd1f291d7ded988e1f) chore: 工单创建，当用户查询为空异常处理
- [`070ad96`](https://github.com/Duke1616/ecmdb/commit/070ad9696c34b4d8b4905c0f45cae8a799546f21) fix: ioc 注册 库名写死，导致切换其余数据库无法启动
- [`dce8a37`](https://github.com/Duke1616/ecmdb/commit/dce8a3713862d75d39fbca8ba01628c88bdd85ae) chore: 优化前端自动化流程提供展示
- [`14d5bda`](https://github.com/Duke1616/ecmdb/commit/14d5bdaa96e3bea1dc94f2f1211931fae94dcdea) chore: 任务历史里面添加代码库名称字段
- [`fdb7702`](https://github.com/Duke1616/ecmdb/commit/fdb770218f41f6bef98a7720be48fdaa7c5fccad) Merge branch 'main' of https://github.com/Duke1616/ecmdb
- [`2091c0c`](https://github.com/Duke1616/ecmdb/commit/2091c0c48b4310311d56283e1b76e169429fd993) 策略

## [v1.6.2](https://github.com/Duke1616/ecmdb/releases/tag/v1.6.2) - 2024-10-11

- [`b009298`](https://github.com/Duke1616/ecmdb/commit/b0092980d2b82ed8b2732c0f41854e8a49598b2d) fix: 替换新的菜单、权限部署SQL文件

## [v1.6.1](https://github.com/Duke1616/ecmdb/releases/tag/v1.6.1) - 2024-10-11

- [`8b4591a`](https://github.com/Duke1616/ecmdb/commit/8b4591a3dd0e5c283a3a0d9c7883c1c86176b273) fix: 调整任务运行顺序，修复 err.Error nil 的情况

## [v1.6.0](https://github.com/Duke1616/ecmdb/releases/tag/v1.6.0) - 2024-10-11

- [`81426bb`](https://github.com/Duke1616/ecmdb/commit/81426bb0245b053a55da510ba868602ca736ec91) chore: 任务状态新增调度失败情况
- [`9b6fd8b`](https://github.com/Duke1616/ecmdb/commit/9b6fd8b16a70dad0ffc70cee0caeb7f8402bf5a1) chore: 优化部分接口性能
- [`7c2a13a`](https://github.com/Duke1616/ecmdb/commit/7c2a13a7452d285baa71beee5858556248bccd24) chore: 删除部分弃用代码，优化函数名称
- [`3558f50`](https://github.com/Duke1616/ecmdb/commit/3558f50018e1b4fa31fbfb713aeea58d078122cd) chore: 模型关联 查询bug、etcd连接异常panic
- [`2e4d81e`](https://github.com/Duke1616/ecmdb/commit/2e4d81e20ae9aa0acff3363a71aec219f82e30d1) chore: 优化企业微信来源审批，消息通知
- [`fba3b6d`](https://github.com/Duke1616/ecmdb/commit/fba3b6d32ea0237574edead28b122210a81429bd) chore: 完善审批通过、驳回状态变更
- [`26a2c3b`](https://github.com/Duke1616/ecmdb/commit/26a2c3b8b123e4f27afbb0ce2ec5387285e3fb72) chore: 工单撤销
- [`488349a`](https://github.com/Duke1616/ecmdb/commit/488349a398e5a9fafc4836e0430d63ef2bb7ea49) chore: 修改查询工单状态数组
- [`995611b`](https://github.com/Duke1616/ecmdb/commit/995611b88aaecf9455fc00bef0636a20124f8369) chore: 工单历史
- [`ce2bdbd`](https://github.com/Duke1616/ecmdb/commit/ce2bdbddf9fa3c5a2218db6d6786e04172184140) chore: 优化工单列表提单人、处理人前端展示名称
- [`f33f3ce`](https://github.com/Duke1616/ecmdb/commit/f33f3ce1103523a67b69ee061eb26cc835bfa7e8) refactor: 重构工单消息通知，新增验证流程控制是否开启消息通知，封装 NotifierIntegration 接口支持多消息接收源
- [`d8c7414`](https://github.com/Duke1616/ecmdb/commit/d8c74147906a6471317f0c4b689382eae55650c4) chore: 前端提供接口支持
- [`c74b6e9`](https://github.com/Duke1616/ecmdb/commit/c74b6e9e5f37f317d61b868659d8bd5d08abceef) feat: 新增属性 link 字段, 支持跳转情况
- [`83303d8`](https://github.com/Duke1616/ecmdb/commit/83303d89302e39e269e70fe006493b9f930fbe78) feat: 完善用户管理、部门管理模版，前端联调
- [`98daea0`](https://github.com/Duke1616/ecmdb/commit/98daea0741364423d9b6debe855c092a55befab9) feat: 新增部分管理模块
- [`5df4ebf`](https://github.com/Duke1616/ecmdb/commit/5df4ebf225d80e9562ae93f0f5609531a287cff5) refactor: 改写 notify 消息通知，使用 notify.NotifierWrap 数组
- [`4b4b810`](https://github.com/Duke1616/ecmdb/commit/4b4b81017b0384b1b81574d1126dcaebff955dd5) feat: 添加CMDB相关，资产、属性修改接口
- [`5b48e4b`](https://github.com/Duke1616/ecmdb/commit/5b48e4b582f9c8d2e09f31ae0fe7f75f4f6b4d85) fix: 升级 golang 版本，dockerfile 打包镜像
- [`4df159c`](https://github.com/Duke1616/ecmdb/commit/4df159cf3c010a2c0e031bf0ea044ff73fe66720) fix: 参数传递错误
- [`cd0f1b5`](https://github.com/Duke1616/ecmdb/commit/cd0f1b55ef7f573fb9eb151d66c65fc04d9bdfcc) fix: 修复 wire 依赖注入
- [`29d485f`](https://github.com/Duke1616/ecmdb/commit/29d485f49cf20b5f17ab7df8fdcee49cee084d0a) feat: 当在飞书点击审批后，撤回消息, 防止显示过乱
- [`7018251`](https://github.com/Duke1616/ecmdb/commit/7018251c00b26e777410f1b6e855c7100c5f68ab) feat: 接入飞书回调，审批通过、拒绝
- [`fd58b79`](https://github.com/Duke1616/ecmdb/commit/fd58b790c07f67891c72d90125479a7cf33ac616) feat: 接入 enotify 消息通知
- [`2d3c5ae`](https://github.com/Duke1616/ecmdb/commit/2d3c5ae1933a383c53f3643282afcc74a1ec1d85) chore: 服务启动美化
- [`5ee4241`](https://github.com/Duke1616/ecmdb/commit/5ee424134223fe55eead2ef38cb7ab2f8bdbd52c) fix: getVersion 逻辑返回错误，修正
- [`11f168c`](https://github.com/Duke1616/ecmdb/commit/11f168ca476f31721e2c61f9198341df02782cb9) fix: Dockerfile 打包镜像传递版本信息
- [`d137736`](https://github.com/Duke1616/ecmdb/commit/d1377361bf9de9695c40933080cc0ad7b7c31714) fix: 版本号 compare 比较

## [v1.5.0](https://github.com/Duke1616/ecmdb/releases/tag/v1.5.0) - 2024-08-26

- [`148048b`](https://github.com/Duke1616/ecmdb/commit/148048bb6d39c3f22307030b8583257b9d53bfe6) fix: 修复第一次查询版本为空的情况
- [`3c7f2bf`](https://github.com/Duke1616/ecmdb/commit/3c7f2bf8d97f9576d78a877ae49e3acc181464dc) feat: 新增 initial 全量增量模式初始化数据方式
- [`95d2919`](https://github.com/Duke1616/ecmdb/commit/95d291978db52b142c013a368085c27a2dc04d34) fix: 添加github action

## [v1.4.1](https://github.com/Duke1616/ecmdb/releases/tag/v1.4.1) - 2024-08-23

- [`4b1d2fb`](https://github.com/Duke1616/ecmdb/commit/4b1d2fbaa7e046f1edd764cb9ea6ace6e4bf8543) fix: 删除bumps
- [`15c63e7`](https://github.com/Duke1616/ecmdb/commit/15c63e7147ab6ca31cb87c0742937fa07b84b57a) fix: upleft release
- [`5496a52`](https://github.com/Duke1616/ecmdb/commit/5496a529cdacf37064f5710efad9401bb3f9e4cd) action
- [`7b37e54`](https://github.com/Duke1616/ecmdb/commit/7b37e545c57a36c2db2d4cc6318e7c60452efc9b) action
- [`0f256f5`](https://github.com/Duke1616/ecmdb/commit/0f256f51a97cedd4babbe383c060c0a892ce8048) chorm: 修改 uplift.yaml

## [v1.4.0](https://github.com/Duke1616/ecmdb/releases/tag/v1.4.0) - 2024-08-23

- [`7031e93`](https://github.com/Duke1616/ecmdb/commit/7031e93d3de57c1132604dca38cf34f47cb293c5) 简单完善一下文档
- [`9c92a13`](https://github.com/Duke1616/ecmdb/commit/9c92a138eeab403b2a8a0bd4aae9310fd161730a) 完善用户登录逻辑
- [`0121aa1`](https://github.com/Duke1616/ecmdb/commit/0121aa1364a69b09c774549299d03b663e76cac2) 部署 compose 编写
- [`d520b62`](https://github.com/Duke1616/ecmdb/commit/d520b62e0e8543076363da17838eac231221c382) fixbug: 创建用户逻辑错误
- [`1861db0`](https://github.com/Duke1616/ecmdb/commit/1861db0daaaeb38b2f0c7f32e2d4dc0d38e0335f) fixbug：修复没有数据的情况，返回错误
- [`6e64c05`](https://github.com/Duke1616/ecmdb/commit/6e64c0550e4bd2aea961ffa4fae6b06bad5075b8) 完善部署文档
- [`93f5550`](https://github.com/Duke1616/ecmdb/commit/93f555046f4e73bf1d90c92cd843f27ff980a6c5) fix: 修复readme.md中安装步骤；增加了部署文档和安装脚本
- [`682c4a2`](https://github.com/Duke1616/ecmdb/commit/682c4a2f23c6e7d1ab7f224ac95b7b1f35df985d) fix: 修复readme.md文件中的IP地址
- [`8a72717`](https://github.com/Duke1616/ecmdb/commit/8a7271775c5fb8582d311b0bce2f225374a36c0a) Merge pull request #3 from cplinux98/fix_install
- [`5d55b3c`](https://github.com/Duke1616/ecmdb/commit/5d55b3c57e63d3863590acc6040a1bdd724516de) cobra 封装命令行启动
- [`eb8c7e5`](https://github.com/Duke1616/ecmdb/commit/eb8c7e5bfead407344b9cf250523ad191e6991af) GOPROXY 代理
- [`1c91695`](https://github.com/Duke1616/ecmdb/commit/1c9169560667af16e369ed6060ec795e0eb5ffbb) fix: 添加 changelog
- [`f944ebb`](https://github.com/Duke1616/ecmdb/commit/f944ebb277998ed2bdccf06eec07065db4a19294) feat: 简单实现，init 初始化数据，版本增量数据，但不支持降级操作
- [`fc914ca`](https://github.com/Duke1616/ecmdb/commit/fc914ca690721ad163011e71709c3866fc68a794) 添加临时

## [v1.3.0](https://github.com/Duke1616/ecmdb/releases/tag/v1.3.0) - 2024-08-15

- [`20e9cf6`](https://github.com/Duke1616/ecmdb/commit/20e9cf640b40bed2f2472a743ec619a3f68f1c2e) readme
- [`6997cfa`](https://github.com/Duke1616/ecmdb/commit/6997cfaaff00ebe59466971328852c74cc4718d5) 增强runner，对敏感变量脱敏
- [`7702d38`](https://github.com/Duke1616/ecmdb/commit/7702d3842fe2da447a822cb7c76dde798d753cbb) 任务变量脱敏
- [`f9708e3`](https://github.com/Duke1616/ecmdb/commit/f9708e33181c41b97a8087aa27ee55f1b64225e2) 变量数据库层面通过 AES 加密存储
- [`af1c40b`](https://github.com/Duke1616/ecmdb/commit/af1c40b51bd91985588a1b8e3d4c641a15446cb5) 配置文件 example 同步
- [`d18391b`](https://github.com/Duke1616/ecmdb/commit/d18391b65ba0947d0113eb6381b0a93b6a2971be) 新增重试状态
- [`18497a3`](https://github.com/Duke1616/ecmdb/commit/18497a3978f305870427f1aafab389bd65587122) fixbug: 逻辑问题，当触发修改时, 加密变量会变更为None
- [`b102824`](https://github.com/Duke1616/ecmdb/commit/b102824a8b99ce2d7e950c0d40f65aab8696cf9b) 权限设计
- [`1481baa`](https://github.com/Duke1616/ecmdb/commit/1481baa22ac8995a2f6f1a0708ce963a531e63ff) 去除casbin配置文件
- [`e127be6`](https://github.com/Duke1616/ecmdb/commit/e127be6ed96edf0240c0a02be17258428c3dae32) 集成 casbin 策略功能
- [`e9bff7e`](https://github.com/Duke1616/ecmdb/commit/e9bff7e1c84ec111ab12ff9aea0bae00338b284c) menu 菜单
- [`c4d29f0`](https://github.com/Duke1616/ecmdb/commit/c4d29f034a6125fc0775ff5232db00e97245d83d) api 注册
- [`8fc2866`](https://github.com/Duke1616/ecmdb/commit/8fc28664f84584b9b874e0e6411103c9c956b7da) 角色模块
- [`7f37e77`](https://github.com/Duke1616/ecmdb/commit/7f37e779ccd7027cd663e78020cdfced87d1ff34) 权限模块联调
- [`3560592`](https://github.com/Duke1616/ecmdb/commit/35605920aa58929ec3c4f8662576b86e8f3c4a36) 语法错误
- [`62e74ff`](https://github.com/Duke1616/ecmdb/commit/62e74ff9b43d44f70586e4a5b89dbc90f000caad) 角色完成 50%，待开发，用户模块进行对接联调
- [`09c59dc`](https://github.com/Duke1616/ecmdb/commit/09c59dc4aebc6fb29b8b4b0a9d0b1f7d56ed00bb) 用户模块
- [`477056d`](https://github.com/Duke1616/ecmdb/commit/477056d5517bd84b8fa4cdc9fe34003cb411478f) 补充部署逻辑，录入用户角色信息到casbin库中
- [`fdec76b`](https://github.com/Duke1616/ecmdb/commit/fdec76bb1a874d5e52a9ff7dae50b9fdaa35af3a) fixbug: casbin 没有 filter gourping的方法
- [`497cded`](https://github.com/Duke1616/ecmdb/commit/497cdedace21b814646a1d5a0123502b0c996b62) 动态权限返回给前端
- [`34ec849`](https://github.com/Duke1616/ecmdb/commit/34ec84917c5ae6b7e03be0b93bc07fc723f4f77c) 抽出 permission 模块，优化菜单与角色之间映射关系
- [`068e0aa`](https://github.com/Duke1616/ecmdb/commit/068e0aa5a2105ac864425f4849411069f2d80ed2) 修改路由名称
- [`67998aa`](https://github.com/Duke1616/ecmdb/commit/67998aa3e53cada125f5e400732d97cd788de226) 优化部分逻辑
- [`3c8fabb`](https://github.com/Duke1616/ecmdb/commit/3c8fabbc66dafc002c2cd2a0ef9ed909e2d1384e) 优化逻辑
- [`7786ba6`](https://github.com/Duke1616/ecmdb/commit/7786ba6186aad6e0801c0f6763dcaca6700e0909) 调整 gin 路由注册顺序
- [`bf519cb`](https://github.com/Duke1616/ecmdb/commit/bf519cbbfefc8bff028b87b1553bf23b55776aa7) fixbug: 获取用户菜单权限，通过sess获取用户id
- [`99f1704`](https://github.com/Duke1616/ecmdb/commit/99f17046fa40a33285b1d1cde0521bb515d3bab2) github action 替换阿里云镜像
- [`46016b6`](https://github.com/Duke1616/ecmdb/commit/46016b606260b1aac6e57ff89a0db847c449ec34) 去除缓存，报错

## [v1.2.0](https://github.com/Duke1616/ecmdb/releases/tag/v1.2.0) - 2024-07-29

- [`bfe93e4`](https://github.com/Duke1616/ecmdb/commit/bfe93e48a336ed64443af0369b14399eaea4a31b) 新增MQ：对接 wechat 回调信息
- [`1ae3e93`](https://github.com/Duke1616/ecmdb/commit/1ae3e9305b16f1f0c0e2e54507ecee5f08ea19cb) 同步wechat OA 模版信息
- [`cbc6b06`](https://github.com/Duke1616/ecmdb/commit/cbc6b06504f7eda6fff5dcaf52b4b786fb3ddbab) 优化处理逻辑，把调用wechat移动到service中
- [`d89bd77`](https://github.com/Duke1616/ecmdb/commit/d89bd77d9373aa42636429d80ede02fcfec0d8c1) 适配前端 form-create 数据录入数据库
- [`481e2e0`](https://github.com/Duke1616/ecmdb/commit/481e2e015a7d0ba852712d3bd638b67b194cf2a7) 工单模版，增加查询接口
- [`7bd7a5c`](https://github.com/Duke1616/ecmdb/commit/7bd7a5c949a5d735a41c901b6de24d9f6e8848a0) 新增 codebook 模块
- [`ea129e1`](https://github.com/Duke1616/ecmdb/commit/ea129e1ec104b1a9f97c0cf6461fd406120284e3) 去除 output
- [`66a028b`](https://github.com/Duke1616/ecmdb/commit/66a028bfaaa5b2e075ceeb4ccc349926c8aa550a) 完善codebook模块CRUD
- [`9e62597`](https://github.com/Duke1616/ecmdb/commit/9e625979ad084aeda4a4c842940a165ec4915024) 新增 worker 工作节点模块
- [`999275f`](https://github.com/Duke1616/ecmdb/commit/999275f9069da54d9b157eb1633d093418af4299) 对接 runner 执行器
- [`6797753`](https://github.com/Duke1616/ecmdb/commit/6797753f3ff498e8f2ac29ebfdd4f49cca195bd8) mongodb 配置 ioc引用
- [`85031d6`](https://github.com/Duke1616/ecmdb/commit/85031d6270bd58f6f25e94d57b7b1412344f4b7f) 修改节点状态
- [`65df10f`](https://github.com/Duke1616/ecmdb/commit/65df10f71cd416791bb24ea1c0491b62b14ba0da) 修改节点状态
- [`f454000`](https://github.com/Duke1616/ecmdb/commit/f454000e113bd3f252ccc396358584cc5fc0552e) README 文档
- [`dc51b6e`](https://github.com/Duke1616/ecmdb/commit/dc51b6e030cb0c51af5d19ec2acbd2fdd0d0c95c) 换行
- [`fdbffdb`](https://github.com/Duke1616/ecmdb/commit/fdbffdb87b2e76195d71f6c3d3ccee99f0dd5e81) 新增 list 接口
- [`53ab63f`](https://github.com/Duke1616/ecmdb/commit/53ab63f5acee00aa85a69ae05ff3bfb1cce2bd27) 节点注册，替换成 ETCD
- [`cc97b9e`](https://github.com/Duke1616/ecmdb/commit/cc97b9e687c2583ee1cac08022579fb6d456f13e) 服务启动，自动开启topic
- [`faf5780`](https://github.com/Duke1616/ecmdb/commit/faf578011c14a37747a40bbcea4861bccd1f9f0c) 优化完成 worker 注册
- [`b8e8afd`](https://github.com/Duke1616/ecmdb/commit/b8e8afd2c9b9c6714e4dd49d0f17dc96d8033737) 启动补偿机制，etcd + mongodb 数据库数据校对
- [`a741029`](https://github.com/Duke1616/ecmdb/commit/a7410295d7c6535e7e5c6b9737b732da8572fa42) runner 注册
- [`9b2bc2a`](https://github.com/Duke1616/ecmdb/commit/9b2bc2a544f96e4e03f52342973906dacb4b556d) 消息队列注册验证
- [`0ee24eb`](https://github.com/Duke1616/ecmdb/commit/0ee24ebfe039c3627358ece228c53a81c5448fea) 工单、策略模块
- [`6f7ebef`](https://github.com/Duke1616/ecmdb/commit/6f7ebef8b42e653d410f1a9016a1ad2acbf273d0) 添加 order 工单信息
- [`9dc0690`](https://github.com/Duke1616/ecmdb/commit/9dc069098d84695ca9123186810600b7687f1c62) 集成 easyflow
- [`2eec347`](https://github.com/Duke1616/ecmdb/commit/2eec347eaa178b8fb61c4c5f42ad95d56412b705) 适配 logic flow
- [`cc75201`](https://github.com/Duke1616/ecmdb/commit/cc75201b72a517b36b2f1b57814693c752cdc1cf) workflow crud
- [`9cbe0f5`](https://github.com/Duke1616/ecmdb/commit/9cbe0f5767d616ba7b0e914f8cc578dea2f5f356) 添加 LICENSE
- [`af79840`](https://github.com/Duke1616/ecmdb/commit/af798405902e0dfe074fac99952415f04689436e) easy-workflow 简单测试
- [`cc56c47`](https://github.com/Duke1616/ecmdb/commit/cc56c47216c736ef375b0a4e0f1b06d2e0bde974) 模版分组
- [`dc9f437`](https://github.com/Duke1616/ecmdb/commit/dc9f4376129d2f890e6222242dc5fab6deeb611e) 新增字段
- [`78c4ef0`](https://github.com/Duke1616/ecmdb/commit/78c4ef0ccca19cefd3889ca460d57769647bf67a) 聚合组数据查询
- [`f113f48`](https://github.com/Duke1616/ecmdb/commit/f113f487067afea3c1ad68b63e031a965b0bead4) order 创建工单，同步本地数据库，发送事件到Kafka
- [`3ff0dae`](https://github.com/Duke1616/ecmdb/commit/3ff0daeb252bda9c8641cf232d0bb7224d3228ca) order 模块
- [`2cbaaa9`](https://github.com/Duke1616/ecmdb/commit/2cbaaa906a4c2ad91f7e2792fe11d452c2a44cfa) 联调前端，创建工单，对接后端流程引擎
- [`fd61893`](https://github.com/Duke1616/ecmdb/commit/fd618938fabc738bbe857e5f38d8f5f718be36a6) elog 日志打印问题
- [`022f91a`](https://github.com/Duke1616/ecmdb/commit/022f91aa8f92261fc3dd3e74805f7791b8a36d18) todo order
- [`3e362c0`](https://github.com/Duke1616/ecmdb/commit/3e362c0ded4cdac2358dc7f7870280903f503fd4) 调整目录结构，流程引擎 engine
- [`a7f0a54`](https://github.com/Duke1616/ecmdb/commit/a7f0a5433f0697eef131a50ff9dc95ae314f4da9) 调整目录组织，解决循环引用，通过Kafka解偶 easyflow event调用order修改状态
- [`d90fbf7`](https://github.com/Duke1616/ecmdb/commit/d90fbf778dfd060a8a7eef867d639118168ceeee) 流程引擎代码，抽象 Instance 统一展示
- [`0506730`](https://github.com/Duke1616/ecmdb/commit/0506730493b8e324f543dfd09571c1e0bad0b452) pass 流程
- [`0d32bb4`](https://github.com/Duke1616/ecmdb/commit/0d32bb48a574234e41027cc3f06f7b3282dca792) 并行网关
- [`13f2fbd`](https://github.com/Duke1616/ecmdb/commit/13f2fbd91eebaa2a59a8a1931bb4d3c4f9e89dce) 包容网关、并行网关
- [`bb370fa`](https://github.com/Duke1616/ecmdb/commit/bb370faf548a11995056ad7f44f89e0a00797bf9) 审批记录
- [`28c9793`](https://github.com/Duke1616/ecmdb/commit/28c979309eff53c996b922b89867e916519eb0b9) 改写获取我的工单列表
- [`16ff3b7`](https://github.com/Duke1616/ecmdb/commit/16ff3b74ed9fd78264f03c0f8bb126fedc51cdbd) 前端 el-table 动态 合并单元格
- [`3eb0465`](https://github.com/Duke1616/ecmdb/commit/3eb0465486d8364c374cc585f5b078718d679641) engine 拆分 event 为独立模块
- [`fad25e8`](https://github.com/Duke1616/ecmdb/commit/fad25e8318d11e12da0ab46f3e5b4aac9f36af7e) wire 注解
- [`07b5e29`](https://github.com/Duke1616/ecmdb/commit/07b5e2971b55be921a4377248965e166007f9fcc) 联动任务模块
- [`b0a582d`](https://github.com/Duke1616/ecmdb/commit/b0a582d69deede20cd1433b3ee89169406623adc) 自动化执行
- [`81bf74d`](https://github.com/Duke1616/ecmdb/commit/81bf74d816f25de7129590bec884fff2e7629f03) 处理任务执行结果
- [`66f2fc3`](https://github.com/Duke1616/ecmdb/commit/66f2fc3def11676a37a22fc46f00eb02cfb4dc00) 流程图展示
- [`fbfdf68`](https://github.com/Duke1616/ecmdb/commit/fbfdf68f60343c472af6218fc777b2f8531d47c3) 任务历史
- [`d434b18`](https://github.com/Duke1616/ecmdb/commit/d434b18b4cf7792d0289a9509564e0bf1c333ab2) Args 传递参数
- [`9a50f16`](https://github.com/Duke1616/ecmdb/commit/9a50f1626a04f12834d43cd2e1d0c3c976f6efd8) 增强任务模块，支持重试、修改参数，runner模块新增环境变量
- [`ac2cbc2`](https://github.com/Duke1616/ecmdb/commit/ac2cbc216d8b7b88fe1f275c0310564cbd2efa9b) 完成基本自动化功能、支持变量
- [`dab8e85`](https://github.com/Duke1616/ecmdb/commit/dab8e856a90fb566c3f5da7729d175f160b5e5d2) task 定时任务
- [`72bc3f2`](https://github.com/Duke1616/ecmdb/commit/72bc3f233759b1eea4ec739cc00b30523afe7170) 定时任务改为 goroutine 启用
- [`3c327f4`](https://github.com/Duke1616/ecmdb/commit/3c327f4543bac3d686cdd19c981dfee18c4db6ab) 新增关闭自动化流程任务定时任务，任务运行，重试机制
- [`84fc9ca`](https://github.com/Duke1616/ecmdb/commit/84fc9ca8082bf0763d9864aec0c3a5aa18110736) 定时任务启动配置
- [`973df3a`](https://github.com/Duke1616/ecmdb/commit/973df3ad06fe0f8a9e3a44063092d8345e34ba2d) 验证兼容 wechat 审批 OA
- [`172bb4e`](https://github.com/Duke1616/ecmdb/commit/172bb4e4369eb55e2da11a79d57a8193794dc4e2) 工单系统流程图

## [v1.1.0](https://github.com/Duke1616/ecmdb/releases/tag/v1.1.0) - 2024-06-07

- [`687089c`](https://github.com/Duke1616/ecmdb/commit/687089c552029dd54c8bf7b52a65b9262b35592a) fixbug: 左侧伸展方向无法展示
- [`8606786`](https://github.com/Duke1616/ecmdb/commit/8606786b4b511f75f744969b7b63f569ac416fc5) add: 新增属性安全模型
- [`20a6713`](https://github.com/Duke1616/ecmdb/commit/20a6713a3d998290387deea407b36e5698bd47ab) 全局搜索 secure 类型展示

## [v1.0.0](https://github.com/Duke1616/ecmdb/releases/tag/v1.0.0) - 2024-05-28

- [`3b2876c`](https://github.com/Duke1616/ecmdb/commit/3b2876ccb3bf3a9292aa5d155d96589faa46a439) Initial commit
- [`31e4b18`](https://github.com/Duke1616/ecmdb/commit/31e4b1804102fb246c9928bed61e5e2e10fb82aa) 初始化buf、taskfile
- [`cb47651`](https://github.com/Duke1616/ecmdb/commit/cb47651c02d1f983283e16262f2ced0c8335e329) 目录结构设计
- [`2e76ae6`](https://github.com/Duke1616/ecmdb/commit/2e76ae62490db46af19f294a61cbb3d80b42aefb) update
- [`a20dc46`](https://github.com/Duke1616/ecmdb/commit/a20dc46627960e4a4b757c5d94fd6aec31a7b794) 项目初始化
- [`ec614f6`](https://github.com/Duke1616/ecmdb/commit/ec614f6464c15aa89df98619c714162ac1249a7b) 项目初始化
- [`2e82c99`](https://github.com/Duke1616/ecmdb/commit/2e82c992b7bb620ed1665461354258f13e3739e7) CMDB 初始化
- [`280e59c`](https://github.com/Duke1616/ecmdb/commit/280e59cae37f4e7a4d818d29816193a15a2cf932) mongo 连接 探测
- [`cdfbbb7`](https://github.com/Duke1616/ecmdb/commit/cdfbbb7356351fe3690dd4428e416a390d6d7b57) mongo 自增ID
- [`9d0c4f6`](https://github.com/Duke1616/ecmdb/commit/9d0c4f68a6197f21ee4a0c0bfac53ffb4d78f42f) 创建模型逻辑
- [`3d242c7`](https://github.com/Duke1616/ecmdb/commit/3d242c7dfaeda0255364843966ecc2c664cfafdd) 初步设计 attribute
- [`0ef3b92`](https://github.com/Duke1616/ecmdb/commit/0ef3b9229eb7f9de400205d6ca3aedde4c625f42) Resource 基本设计
- [`c06120c`](https://github.com/Duke1616/ecmdb/commit/c06120ca3584867e5f076fefdfc011b6c7685b3d) Resouce 数据录入
- [`623426c`](https://github.com/Duke1616/ecmdb/commit/623426ca233b4865cfa35917ea2839e71e7323b3) resource
- [`3f71a3d`](https://github.com/Duke1616/ecmdb/commit/3f71a3d5863be68071bcbddf2d0464a12f08f45b) model list detail 逻辑
- [`4cd67bd`](https://github.com/Duke1616/ecmdb/commit/4cd67bd9ffda38946b7574e6be1b4a9c87c9f0cb) relation 关联关系
- [`44cc007`](https://github.com/Duke1616/ecmdb/commit/44cc007ed0fe475c4ce5c7fadc3cdf9f38865a1b) wire 依赖
- [`c865c74`](https://github.com/Duke1616/ecmdb/commit/c865c7430c1ec696083916587e5c5157cfe9fb62) 修改数据库存储结构
- [`50943fb`](https://github.com/Duke1616/ecmdb/commit/50943fb9e1558ae223bc67102fd606f42712798b) 修改数据库存储结构
- [`30b0093`](https://github.com/Duke1616/ecmdb/commit/30b00938e904b08a90b103b9607ffcec3d80887b) 待完成字段映射，查询Mongo
- [`0e9ef4c`](https://github.com/Duke1616/ecmdb/commit/0e9ef4c8d717565cf13c7e29478416cc8a594f65) resource 查询逻辑
- [`fc28510`](https://github.com/Duke1616/ecmdb/commit/fc285108dab7145a44a009672d572075c7364ba6) 优化resource 和 attribute 关联处理逻辑
- [`4e01c57`](https://github.com/Duke1616/ecmdb/commit/4e01c57dcbb2defbbdc3685134f84077a695a7f9) 封装 mongox 自增ID
- [`a8ec82f`](https://github.com/Duke1616/ecmdb/commit/a8ec82f2b3cc8a87d63c33ad798df70822304210) 资产关联关系
- [`e787bab`](https://github.com/Duke1616/ecmdb/commit/e787babcc429eca812b37495727c66b665aec4a7) gin context 封装
- [`eef90c0`](https://github.com/Duke1616/ecmdb/commit/eef90c0bdcc4b58972de116b47cf8114024a98ea) 条件查询
- [`f205018`](https://github.com/Duke1616/ecmdb/commit/f20501879eab99aab2c9da08da75ec2547349609) UniqueIdentifier => uid 修改统一命名，代表唯一标识
- [`8f892a7`](https://github.com/Duke1616/ecmdb/commit/8f892a70655d6f553ddb45e445b8347d742a5c02) 前端传递
- [`f8a52dc`](https://github.com/Duke1616/ecmdb/commit/f8a52dc7d221fe0782e58a0e76c68ad5b1b88a01) 获取关联resource数据
- [`e9ddd9a`](https://github.com/Duke1616/ecmdb/commit/e9ddd9a539a4952d67f3cd589089d305cebd52f1) 完善 ioc
- [`51fa4b2`](https://github.com/Duke1616/ecmdb/commit/51fa4b2f876a1bbae4c9489f2b3b898209964194) 新增å通过关联类型和模型UID，查询数资源数据
- [`3df6020`](https://github.com/Duke1616/ecmdb/commit/3df6020c714966bebd51e53c5a280c1436ef0929) 拆分relation为多个文件
- [`8bd5acd`](https://github.com/Duke1616/ecmdb/commit/8bd5acd7731bb34037dd4e30426ed2b02b542a20) realtion type
- [`268df1a`](https://github.com/Duke1616/ecmdb/commit/268df1a0a4bcf5293a733885f7fa00407894185c) 关联类型
- [`f622306`](https://github.com/Duke1616/ecmdb/commit/f622306907fa87c637fb41533fd75db4a961eea7) 模型拓补图
- [`e3915c1`](https://github.com/Duke1616/ecmdb/commit/e3915c151c3e03fe7b26a0f7842e2a3155d5c269) 模型拓补图
- [`14b0df5`](https://github.com/Duke1616/ecmdb/commit/14b0df5324297591ca45ffd9f7062953bb573e65) 封装LDAPX
- [`4fcb08b`](https://github.com/Duke1616/ecmdb/commit/4fcb08bafddd938ec1e0654a84540211c2666561) 用户登录逻辑
- [`d0c22d9`](https://github.com/Duke1616/ecmdb/commit/d0c22d9af3ef33582ad491a28d3a22adb696941b) 用户LDAP登录逻辑
- [`1bc168d`](https://github.com/Duke1616/ecmdb/commit/1bc168dca21bd516cfd7e910800c66469e2163f7) Session + Jwt 登录认证
- [`60878de`](https://github.com/Duke1616/ecmdb/commit/60878dedc89415b3235ef533f23f41e440950674) 继续完善 关联关系模块
- [`4e18851`](https://github.com/Duke1616/ecmdb/commit/4e18851783d7bcaa21b5a8d137360b800c05c712) 通过聚合，处理资源列表
- [`0a8c492`](https://github.com/Duke1616/ecmdb/commit/0a8c4924918e291e803cbcf9d055f4581fd2eecb) 修复聚合列表
- [`88963d6`](https://github.com/Duke1616/ecmdb/commit/88963d6fe0e8f7986d3832ba1013abca5c5ca215) 模型重构
- [`b831eac`](https://github.com/Duke1616/ecmdb/commit/b831eac0f02e960958821208775a1e00cb26e4a2) 资产列表
- [`526f58d`](https://github.com/Duke1616/ecmdb/commit/526f58dc5acb01115d8eafe7ecd110cede2ca49e) 字段添加 detail 方法
- [`45c6ccf`](https://github.com/Duke1616/ecmdb/commit/45c6ccf06001ec2bded9ce4c75288c7172f46ef5) 模型属性 列表
- [`cb30b9e`](https://github.com/Duke1616/ecmdb/commit/cb30b9e511c85026fb1adbc9433e86e12772ca39) 去除没必要的指针
- [`d7d94af`](https://github.com/Duke1616/ecmdb/commit/d7d94af14dc7aa4ef5969bb814f1350698e48500) 真的令人头大
- [`4cf0412`](https://github.com/Duke1616/ecmdb/commit/4cf041221519496e4be93330f762032803c11bec) relation resource 拓补图 资产标记
- [`3183352`](https://github.com/Duke1616/ecmdb/commit/3183352255739d423e6a54f9eb6d6148e3b4924c) update
- [`4a50b32`](https://github.com/Duke1616/ecmdb/commit/4a50b32d7bce903393dbdb175452c56ec788b80c) 重构 relation 创建逻辑
- [`86c7ba0`](https://github.com/Duke1616/ecmdb/commit/86c7ba08145a34cfa871c8d114b0f0ff72edfa5a) 查询可以æ关联的模型数据
- [`7ee2180`](https://github.com/Duke1616/ecmdb/commit/7ee2180f2218924edbb06e408062a090cba100f9) 解决循环引用
- [`1482273`](https://github.com/Duke1616/ecmdb/commit/14822737764265038cded2afc175535824d4fcb1) 测试完成 查询å以关联的节点
- [`3538cc6`](https://github.com/Duke1616/ecmdb/commit/3538cc6763ef426d756bc79087ef5314df3d663f) 计划封装 mongox
- [`e45e02c`](https://github.com/Duke1616/ecmdb/commit/e45e02c53225a75c31a28eb69bdec26a217b5a73) 计划封装 mongox
- [`103d44d`](https://github.com/Duke1616/ecmdb/commit/103d44d89520cbcd29d0e90f41ed667cb66dc349) 修改mongox 作为入参，方便编写测试
- [`ed77274`](https://github.com/Duke1616/ecmdb/commit/ed772747f4b5151aef89041eac22ec8a3eb6551a) e2e 测试
- [`19d7974`](https://github.com/Duke1616/ecmdb/commit/19d7974ff6e8fe543c96fca805c327b6eff06fb3) 创建资源，e2e测试
- [`382179d`](https://github.com/Duke1616/ecmdb/commit/382179df17dcdb6aacbdec828c225133f3454a37) 新增 e2e测试
- [`163c2a3`](https://github.com/Duke1616/ecmdb/commit/163c2a3639e43941905b6afb35877013237af875) 重构 searchattrubute 返回信息
- [`7d37d09`](https://github.com/Duke1616/ecmdb/commit/7d37d095213d462c0490a2bbdaef54e1c795e845) attribute 添加联合唯一索引
- [`2347767`](https://github.com/Duke1616/ecmdb/commit/2347767632b58255ea30d063ef9cfd81222f89da) 优化 attribute 模块，完善e2e测试
- [`97e93cf`](https://github.com/Duke1616/ecmdb/commit/97e93cf06dd03672167db922061a256548d7b9cb) 创建 开启mongo事务
- [`50fd743`](https://github.com/Duke1616/ecmdb/commit/50fd7433c243cce03b44bfd675d218e9db1804f2) 删除多余方法
- [`921d9a8`](https://github.com/Duke1616/ecmdb/commit/921d9a857dbf1ca2bd5ec334cd5e5e7e65a780b5) model 重构
- [`f3c0599`](https://github.com/Duke1616/ecmdb/commit/f3c0599c8656c5fa28a672940b62451f89726433) 循环引用，我吐了
- [`5cb21fd`](https://github.com/Duke1616/ecmdb/commit/5cb21fde53c12a035e6c594ef3200d171d650834) 解决循环引用
- [`64f4418`](https://github.com/Duke1616/ecmdb/commit/64f4418d11f118b8ed76d0bc25b5232de215ae73) 优化 模型关联关系
- [`f04e434`](https://github.com/Duke1616/ecmdb/commit/f04e4345577c58f8bf1ae7c4136eccd9d090c480) 关联类型 优化
- [`634b519`](https://github.com/Duke1616/ecmdb/commit/634b519755eb896e9d1f8fbf511a370df9b2b872) 优化rsource 模块
- [`11c22e7`](https://github.com/Duke1616/ecmdb/commit/11c22e7e5cd79b99b26f6af06039860bc944cdb6) 优化 resource
- [`b019d20`](https://github.com/Duke1616/ecmdb/commit/b019d20d709f20bfc1c9b11db71a2fc39095ea34) 优化
- [`2bf03ba`](https://github.com/Duke1616/ecmdb/commit/2bf03ba4ebe252935a5ddb9d52e1fae0f303eba2) realtion resource 优化完成
- [`de7705c`](https://github.com/Duke1616/ecmdb/commit/de7705c6508a2da65f0ae66060f7a4fc6b1a77ac) user 模块
- [`a723e24`](https://github.com/Duke1616/ecmdb/commit/a723e24d43590e8470dcacb555f0732a001627ad) relation e2e测试模版
- [`96ce2e3`](https://github.com/Duke1616/ecmdb/commit/96ce2e35fdc7e89a1dca17c64fa8bee2b0b35694) 结构体修改，启动错误
- [`79d3721`](https://github.com/Duke1616/ecmdb/commit/79d3721b13c89262aebdca75984823a3d945260b) 模型分组返回
- [`70cd366`](https://github.com/Duke1616/ecmdb/commit/70cd366776547032df934a4ad2cc6f42396d10d0) 去除事务操作
- [`3747bd5`](https://github.com/Duke1616/ecmdb/commit/3747bd5d0646104ec9d56029b163a99a072717f7) 自定义模型列
- [`6d0c93e`](https://github.com/Duke1616/ecmdb/commit/6d0c93e401eb9ab8f05b2e50a10c6de71ef231c0) 前后端联调，模块
- [`00d59f5`](https://github.com/Duke1616/ecmdb/commit/00d59f523e6eb9d7e63b82d4032f9d75d2f21af9) 模型模块、前端联调
- [`4e9a727`](https://github.com/Duke1616/ecmdb/commit/4e9a7279b4a5cd128440fe80a24b64c724ffaa5e) 对接前端
- [`7797167`](https://github.com/Duke1616/ecmdb/commit/77971676c318483dd84f11f3db09cc221e97041e) 新建关联
- [`7017757`](https://github.com/Duke1616/ecmdb/commit/70177571b8a260d93938a23079f5c7622a7a6139) 新增关联，SRC方向过滤查询bug
- [`479592d`](https://github.com/Duke1616/ecmdb/commit/479592d5675cc99ff886a6f0b024681ac3f454ff) 资产 关联信息展示
- [`09c6886`](https://github.com/Duke1616/ecmdb/commit/09c6886195cc1fc605dcbd46f58fcba260836cf6) 取消关联
- [`1501fc0`](https://github.com/Duke1616/ecmdb/commit/1501fc0411a4b4d683107cfc3b2a1600db27cf4c) 取消关联
- [`629e623`](https://github.com/Duke1616/ecmdb/commit/629e623401f273d3424e134b645eaf9f7290433b) 资产拓扑图
- [`06d9261`](https://github.com/Duke1616/ecmdb/commit/06d92617009caa0e8a13d78dfa16202bd7aa2056) 新增模型，初始化创建属性分组、属性字段
- [`e4216aa`](https://github.com/Duke1616/ecmdb/commit/e4216aaae7dd881eb93f5220e8e6f0ae5daa71f8) 模型属性分组
- [`fb4da5e`](https://github.com/Duke1616/ecmdb/commit/fb4da5e74ec4144bbf85463bcf95ea3df5ac2225) 准备测试整体功能性是否完善
- [`f111bed`](https://github.com/Duke1616/ecmdb/commit/f111bed476c053e58c659ae2abf5330a3e4b44bd) 添加 api 前缀
- [`1cea9e6`](https://github.com/Duke1616/ecmdb/commit/1cea9e645fd3a1326b478d8e47eb10e14e6ab23c) 修改 Dockerfile
- [`7aa20ea`](https://github.com/Duke1616/ecmdb/commit/7aa20ea1b1cc74e22efa47bf8dba08206d472368) 部署文件修改，挂载配置文件
- [`84a1171`](https://github.com/Duke1616/ecmdb/commit/84a1171a814dd01d399ad36c2c12fbe0671e4828) github active
- [`445be77`](https://github.com/Duke1616/ecmdb/commit/445be7725a7154682643b7809bc3daaad410f08c) 去除缓存
- [`53b4b80`](https://github.com/Duke1616/ecmdb/commit/53b4b80cea6e3b5201facfc40980ff09827249a8) update
- [`9a007fe`](https://github.com/Duke1616/ecmdb/commit/9a007fe1dda378efb32cd85f4b338a1938673735) 模型删除验证
- [`cbcf21e`](https://github.com/Duke1616/ecmdb/commit/cbcf21e489e9bdaa09c649e4b967396c4cba1a4f) 全局搜索功能
- [`e12424f`](https://github.com/Duke1616/ecmdb/commit/e12424fd4b8dacfa4ddb72e89a942f1d8511fb79) 新增关联，新增过滤条件
- [`253c2e2`](https://github.com/Duke1616/ecmdb/commit/253c2e28b18fcc41713f13e92520479ba812d763) fixbug： 计算count 传递filter输入错误
- [`13af8d4`](https://github.com/Duke1616/ecmdb/commit/13af8d4f438758e3d0427e36c7806d769f45be8c) 资产关联拖布图 left right 方向扩展
