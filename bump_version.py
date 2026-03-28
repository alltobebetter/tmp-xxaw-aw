import sys
import re
import os
import subprocess

def update_version(file_path, pattern, replacement):
    if not os.path.exists(file_path):
        print(f"⚠️ Warning: File not found: {file_path}")
        return False
    
    with open(file_path, 'r', encoding='utf-8') as f:
        content = f.read()
    
    new_content, count = re.subn(pattern, replacement, content)
    
    if count > 0:
        with open(file_path, 'w', encoding='utf-8') as f:
            f.write(new_content)
        print(f"✅ Updated {count} occurrence(s) in {os.path.basename(file_path)}")
        return True
    else:
        print(f"⚠️ Warning: Could not find match for pattern in {os.path.basename(file_path)}")
        return False

def main():
    if len(sys.argv) != 2:
        print("Usage: python bump_version.py <new_version>")
        print("Example: python bump_version.py 2.1.0")
        sys.exit(1)
        
    new_version = sys.argv[1].strip()
    
    # 验证版本号格式如 x.y.z
    if not re.match(r'^\d+\.\d+\.\d+$', new_version):
        print("❌ Error: Version must be in format x.y.z (e.g. 2.1.0)")
        sys.exit(1)

    print(f"🚀 Start bumping version to {new_version}...\n")

    base_dir = os.path.dirname(os.path.abspath(__file__))

    tasks = [
        {
            "desc": "Wails config (wails.json)",
            "file": "wails.json",
            "pattern": r'("productVersion":\s*")\d+\.\d+\.\d+(")',
            "replacement": rf'\g<1>{new_version}\g<2>',
            "repo": "global"
        },
        {
            "desc": "Inno Setup config (TraeProxy.iss) - AppVersion",
            "file": "TraeProxy.iss",
            "pattern": r'(AppVersion=)\d+\.\d+\.\d+',
            "replacement": rf'\g<1>{new_version}',
            "repo": "global"
        },
        {
            "desc": "Inno Setup config (TraeProxy.iss) - OutputBaseFilename",
            "file": "TraeProxy.iss",
            "pattern": r'(OutputBaseFilename=TraeProxy_Setup_v)\d+\.\d+\.\d+',
            "replacement": rf'\g<1>{new_version}',
            "repo": "global"
        },
        {
            "desc": "Frontend UI (frontend/src/App.vue)",
            "file": "frontend/src/App.vue",
            "pattern": r"(const CURRENT_APP_VERSION = ')\d+\.\d+\.\d+(')",
            "replacement": rf'\g<1>{new_version}\g<2>',
            "repo": "global"
        },
        {
            "desc": "Website UI (website/src/app/page.tsx)",
            "file": "website/src/app/page.tsx",
            "pattern": r"(const version = ')\d+\.\d+\.\d+(';)",
            "replacement": rf'\g<1>{new_version}\g<2>',
            "repo": "website"
        },
        {
            "desc": "Website API (website/src/app/api/version/route.ts)",
            "file": "website/src/app/api/version/route.ts",
            "pattern": r"(const version = ')\d+\.\d+\.\d+(';)",
            "replacement": rf'\g<1>{new_version}\g<2>',
            "repo": "website"
        }
    ]

    success = True
    global_files_changed = []
    
    for task in tasks:
        print(f"[{task['desc']}]")
        file_path = os.path.join(base_dir, task["file"].replace("/", os.sep))
        updated = update_version(file_path, task["pattern"], task["replacement"])
        if not updated:
            success = False
        elif task.get("repo") == "global":
            # 记录被修改过的全局仓库文件
            if task["file"] not in global_files_changed:
                global_files_changed.append(task["file"])
        print("-" * 40)

    if success:
        print("\n🎉 All version bumps completed successfully!")
        
        # 自动执行全局 git commit
        if global_files_changed:
            print("\n📦 Automating global git commit...")
            try:
                # 仅将刚刚被脚本修改的全局文件添加到暂存区
                for f in global_files_changed:
                    subprocess.run(["git", "add", f], cwd=base_dir, check=True)
                
                # 执行本地提交 (不推送到远程)
                commit_msg = f"chore(release): bump version to {new_version}"
                subprocess.run(["git", "commit", "-m", commit_msg], cwd=base_dir, check=True)
                
                print(f"✅ Git commit succeeded: {commit_msg}")
            except subprocess.CalledProcessError as e:
                print(f"⚠️ Git commit failed. Error: {e}")
            except Exception as e:
                print(f"⚠️ Unexpected error with Git operation: {e}")
        else:
            print("\n⚠️ No global files were changed, skipping git commit.")
            
    else:
        print("\n⚠️ Finished with some warnings. Please check the logs above.")

if __name__ == "__main__":
    main()
