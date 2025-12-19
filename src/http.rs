use rquickjs::{Ctx, Function};
use reqwest::Client;
use serde_json::{Map, Value as SerdeValue};
use std::sync::Arc;

// 创建一个Arc包装的Client，以便在多个调用间共享
static CLIENT: once_cell::sync::Lazy<Arc<Client>> = once_cell::sync::Lazy::new(|| {
    Arc::new(Client::new())
});

// 实现fetch函数，支持GET和POST方法
pub async fn fetch(url: String, options: Option<Map<String, SerdeValue>>) -> Result<Map<String, SerdeValue>, String> {
    let client = CLIENT.clone();
    let mut request_builder = client.get(&url);
    
    // 处理options参数
    if let Some(opts) = options {
        // 处理method
        if let Some(method) = opts.get("method").and_then(|v| v.as_str()) {
            match method.to_uppercase().as_str() {
                "GET" => request_builder = client.get(&url),
                "POST" => request_builder = client.post(&url),
                "PUT" => request_builder = client.put(&url),
                "DELETE" => request_builder = client.delete(&url),
                _ => return Err(format!("Unsupported HTTP method: {}", method)),
            }
        }
        
        // 处理headers
        if let Some(SerdeValue::Object(headers)) = opts.get("headers") {
            for (key, value) in headers {
                if let Some(value_str) = value.as_str() {
                    request_builder = request_builder.header(key, value_str);
                }
            }
        }
        
        // 处理body
        if let Some(body) = opts.get("body") {
            request_builder = request_builder.json(body);
        }
    }
    
    // 发送请求
    match request_builder.send().await {
        Ok(response) => {
            let status = response.status().as_u16();
            let mut headers_obj = Map::new();
            for (k, v) in response.headers() {
                headers_obj.insert(k.as_str().to_string(), SerdeValue::String(v.to_str().unwrap_or("").to_string()));
            }
            
            let text = match response.text().await {
                Ok(t) => t,
                Err(e) => return Err(format!("Failed to read response body: {:?}", e)),
            };
            
            // 尝试解析为JSON，如果失败则作为字符串返回
            let data = match serde_json::from_str(&text) {
                Ok(json) => json,
                Err(_) => SerdeValue::String(text),
            };
            
            let mut result = Map::new();
            result.insert("status".to_string(), SerdeValue::Number(serde_json::Number::from(status)));
            result.insert("headers".to_string(), SerdeValue::Object(headers_obj));
            result.insert("data".to_string(), data);
            
            Ok(result)
        },
        Err(e) => Err(format!("HTTP request failed: {:?}", e)),
    }
}

// 注册fetch函数到QuickJS上下文
pub fn register_fetch(ctx: &Ctx) -> Result<(), rquickjs::Error> {
    // 直接注册fetch函数，不使用JavaScript包装
    fn fetch_impl<'js>(ctx: Ctx<'js>, url: String) -> Result<rquickjs::Value<'js>, rquickjs::Error> {
        // 创建一个Tokio运行时来执行异步操作
        let rt = tokio::runtime::Runtime::new().unwrap();
        
        // 执行fetch请求
        let result = rt.block_on(fetch(url, None));
        
        match result {
            Ok(map) => {
                // 将Map转换为JSON字符串，然后解析为JS对象
                let json_str = serde_json::to_string(&map).map_err(|e| {
                    rquickjs::Error::new_into_js_message("Map", "JSON string", e.to_string())
                })?;
                
                // 使用eval来创建JS对象
                let js_obj = ctx.eval(format!("JSON.parse('{}')", json_str.replace("'", "\\'")))?;
                Ok(js_obj)
            },
            Err(err) => {
                // 抛出JS错误
                Err(rquickjs::Error::new_into_js_message("fetch", "JS value", err))
            },
        }
    }
    
    // 创建一个全局的fetch函数
    let fetch_func = Function::new(ctx.clone(), fetch_impl)?;
    
    // 获取全局对象
    let global = ctx.globals();
    // 设置fetch函数到全局对象
    global.set("fetch", fetch_func)?;
    
    Ok(())
}