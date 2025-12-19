use rquickjs::{Context, Function, Runtime};

fn add(a: f64, b: f64) -> f64 {
    a + b
}

fn main() {
    // 创建运行时和上下文
    let rt = Runtime::new().unwrap();
    let ctx = Context::full(&rt).unwrap();

    // 在上下文中执行
    ctx.with(|ctx| {
        // 定义一个 Rust 函数：接收两个数字，返回它们的和
        let rust_add = Function::new(ctx.clone(), add).unwrap();

        // 将函数绑定到全局对象（globalThis）
        let global = ctx.globals();
        global.set("rustAdd", rust_add).unwrap();

        // 执行一段 JavaScript 调用这个函数
        let result: f64 = ctx
            .eval(
                r#"
            rustAdd(10, 20);
        "#,
            )
            .unwrap();

        println!("JS result: {}", result); // 输出: JS result: 30
    });
}
