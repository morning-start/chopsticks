use rquickjs::{Context, Runtime};

// 导入http模块
mod http;

fn main() {
    // 创建运行时和上下文
    let rt = Runtime::new().unwrap();
    let ctx = Context::full(&rt).unwrap();

    // 在上下文中执行
    ctx.with(|ctx| {
        // 注册fetch函数到全局对象
        http::register_fetch(&ctx).unwrap();

        // 执行一段 JavaScript 调用fetch函数，测试GET请求
        let result: String = ctx
            .eval(
                r#"
            // 测试GET请求
            const getResponse = fetch('https://httpbin.org/get');
            JSON.stringify(getResponse);
        "#,
            )
            .unwrap();

        println!("JS result: {}", result);
    });
}
