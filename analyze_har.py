import json
import sys

def parse_har(file_path):
    try:
        with open(file_path, 'r', encoding='utf-8-sig') as f:
            data = json.load(f)
        
        entries = data.get('log', {}).get('entries', [])
        found_count = 0
        
        for entry in entries:
            req = entry.get('request', {})
            url = req.get('url', '')
            
            # 过滤出 OpenAI 或 Anthropic 的请求，或者是常见的 completions 路径
            if 'openai' in url or 'anthropic' in url or '/chat/completions' in url:
                found_count += 1
                print("="*60)
                print(f"【{found_count}】请求目标: [{req.get('method')}] {url}")
                
                print("\n[核心 Headers]")
                for h in req.get('headers', []):
                    # 只打印核心 Header，过滤掉不需要关注的 Cookie 或冗长系统头
                    name = h['name'].lower()
                    if name in ['authorization', 'api-key', 'content-type', 'user-agent', 'x-trae-version', 'host']:
                        # 脱敏处理 Authorization
                        val = h['value']
                        if name == 'authorization' and len(val) > 15:
                            val = val[:10] + '...' + val[-5:]
                        print(f"  - {h['name']}: {val}")
                        
                post_data = req.get('postData', {})
                if post_data:
                    print("\n[Body Payload (部分展示)]")
                    text = post_data.get('text', '')
                    try:
                        parsed = json.loads(text)
                        print(f"  - 包含的字段: {list(parsed.keys())}")
                        print(f"  - 请求的模型 (Model): {parsed.get('model', 'N/A')}")
                        if 'messages' in parsed:
                            print(f"  - 携带的消息数量: {len(parsed['messages'])}")
                        if 'stream' in parsed:
                            print(f"  - 是否流式响应 (Stream): {parsed['stream']}")
                    except Exception:
                        print(f"  - 非标准的 JSON，前150个字符:\n    {text[:150]}...")
                print("="*60)
                
        if found_count == 0:
            print("在这个 HAR 文件中没有发现明文请求 OpenAI 或特殊聊天接口的行为。有可能加密了，或者走的完全不同域名的私有接口。")
            
    except Exception as e:
        print(f"文件解析失败: {e}")

if __name__ == '__main__':
    parse_har('Trae-CustomModel.har')
