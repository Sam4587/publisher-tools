# 归档文件说明

本目录包含已废弃的文件,仅作备份用途。

## 归档时间
2026-02-22

## 归档文件列表

### 启动脚本
- `start.bat` - 旧的启动脚本,已被 `start-all.bat` 替代
- `start-stable.bat` - 使用service-manager的启动脚本,已被 `start-all.bat` 替代
- `service-manager.js` - 服务管理器,已被 `start-all.bat` + `run_hidden.js` 替代

### VBS启动器
- `启动服务.vbs` - VBS启动器,引用了已废弃的 `start.bat`
- `停止服务.vbs` - VBS停止器,功能与直接双击 `stop.bat` 重复

### 服务器文件
- `test-server.js` - 测试服务器,未使用

## 归档原因

### VBS文件归档原因
1. **功能重复**: 用户可以直接双击 `.bat` 文件
2. **维护成本**: 需要保持VBS和BAT文件的同步
3. **已失效**: `启动服务.vbs` 引用的 `start.bat` 已不存在
4. **用户习惯**: 现代用户习惯直接使用 `.bat` 文件

## 当前推荐使用

- 启动服务: `start-all.bat` (直接双击)
- 停止服务: `stop.bat` (直接双击)
- 健康检查: `health-check.bat` (直接双击)

## 注意

这些文件已不再维护,如需恢复请谨慎使用。
