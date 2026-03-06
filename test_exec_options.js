// 测试 exec 模块的 options 参数支持
const result1 = exec.exec("echo", ["Hello", "World"]);
if (result1.success) {
    log.info("测试1 - 基本执行: " + result1.stdout);
} else {
    log.error("测试1 失败: " + result1.error);
}

// 测试2 - 使用 array 类型的 args
const result2 = exec.exec("echo", ["Test with array"]);
if (result2.success) {
    log.info("测试2 - array args: " + result2.stdout);
} else {
    log.error("测试2 失败: " + result2.error);
}

// 测试3 - 使用 options 参数
const result3 = exec.exec("echo", ["Test with options"], {
    timeout: 5000
});
if (result3.success) {
    log.info("测试3 - with options: " + result3.stdout);
} else {
    log.error("测试3 失败: " + result3.error);
}

// 测试4 - shell 命令带 options
const result4 = exec.shell("echo Shell test", {
    timeout: 5000
});
if (result4.success) {
    log.info("测试4 - shell with options: " + result4.stdout);
} else {
    log.error("测试4 失败: " + result4.error);
}

// 测试5 - powershell 命令带 options
const result5 = exec.powershell("Write-Host 'PowerShell test'", {
    timeout: 5000
});
if (result5.success) {
    log.info("测试5 - powershell with options: " + result5.stdout);
} else {
    log.error("测试5 失败: " + result5.error);
}

log.info("所有测试完成！");
