# QuickJS 注入 Rust 函数使用指南

## 基本步骤

### 1. 添加依赖
在 `Cargo.toml` 中添加 `rquickjs` 依赖：
```toml
[dependencies]
rquickjs = "0.2.0"
```

### 2. 编写 Rust 函数
定义需要暴露给 JavaScript 的 Rust 函数，参数和返回值需符合 rquickjs 支持的类型：
```rust
fn add(a: f64, b: f64) -> f64 {
    a + b
}
```

### 3. 创建运行时和上下文
```rust
let rt = Runtime::new().unwrap();
let ctx = Context::full(&rt).unwrap();
```

### 4. 包装并绑定函数
```rust
ctx.with(|ctx| {
    
    // 绑定到全局对象
    let global = ctx.globals();
    // 绑定函数
    global.set("rustAdd", Function::new(ctx.clone(), add)).unwrap();
    
    // 执行 JS 代码调用 Rust 函数，获取返回值。
    // 和绑定的函数一致
    let result: f64 = ctx.eval(r#"rustAdd(10, 20);"#).unwrap();
    println!("JS result: {}", result);
});
```

## 完整示例

```rust
use rquickjs::{Context, Function, Runtime};
use std::error::Error;

fn add(a: f64, b: f64) -> f64 {
    a + b
}

fn main() -> Result<(), Box<dyn Error>> {
    let rt = Runtime::new()?;
    let ctx = Context::full(&rt)?;

    ctx.with(|ctx| {
        let rust_add = Function::new(ctx.clone(), add)?;
        let global = ctx.globals();
        global.set("rustAdd", rust_add)?;

        let result: f64 = ctx.eval(r#"rustAdd(10, 20);"#)?;
        println!("JS result: {}", result);
        Ok(())
    })
}
```

## 更多类型示例

### 字符串处理
```rust
fn greet(name: &str) -> String {
    format!("Hello, {}!", name)
}

// 在 ctx.with 中绑定
let greet_func = Function::new(ctx.clone(), greet)?;
global.set("greet", greet_func)?;

// JS 调用
let greeting: String = ctx.eval(r#"greet('World')"#)?;
println!("{}", greeting); // 输出: Hello, World!
```

### 布尔值与无返回值
```rust
fn is_even(n: i32) -> bool {
    n % 2 == 0
}

fn log(message: &str) {
    println!("JS Log: {}", message);
}

// 绑定
let is_even_func = Function::new(ctx.clone(), is_even)?;
global.set("isEven", is_even_func)?;

let log_func = Function::new(ctx.clone(), log)?;
global.set("log", log_func)?;
```