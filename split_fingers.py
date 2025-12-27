#!/usr/bin/env python3
"""
指纹 YAML 文件拆分脚本
将包含多个指纹的 YAML 文件拆分成单个指纹一个文件
"""

import os
import yaml
import re
from pathlib import Path


def sanitize_filename(name):
    """
    将指纹名称转换为安全的文件名
    移除特殊字符，替换为下划线
    """
    # 移除或替换不安全的字符
    name = re.sub(r'[<>:"/\\|?*]', '_', name)
    # 移除多余的空格
    name = name.strip()
    # 替换空格为下划线
    name = name.replace(' ', '_')
    return name


def split_yaml_file(input_file):
    """
    拆分单个 YAML 文件

    Args:
        input_file: 输入的 YAML 文件路径
    """
    input_path = Path(input_file)
    print(f"处理文件: {input_file}")

    # 读取 YAML 文件
    try:
        with open(input_file, 'r', encoding='utf-8') as f:
            data = yaml.safe_load(f)
    except Exception as e:
        print(f"  错误: 无法读取文件 - {e}")
        return

    if not isinstance(data, list):
        print(f"  警告: 文件格式不是列表，跳过")
        return

    # 创建输出目录（使用原文件名作为目录名）
    # 例如: fingers/http/cdn.yaml -> fingers/http/cdn/
    output_dir = input_path.parent / input_path.stem
    os.makedirs(output_dir, exist_ok=True)

    # 统计信息
    total = len(data)
    success = 0

    # 遍历每个指纹
    for idx, finger in enumerate(data, 1):
        if not isinstance(finger, dict) or 'name' not in finger:
            print(f"  警告: 第 {idx} 个条目没有 name 字段，跳过")
            continue

        finger_name = finger['name']
        safe_name = sanitize_filename(finger_name)

        # 生成输出文件名
        output_file = output_dir / f"{safe_name}.yaml"

        # 如果文件已存在，添加序号
        if output_file.exists():
            counter = 1
            while (output_dir / f"{safe_name}_{counter}.yaml").exists():
                counter += 1
            output_file = output_dir / f"{safe_name}_{counter}.yaml"

        # 写入单个指纹文件（保持为列表格式，只包含一个元素）
        try:
            with open(output_file, 'w', encoding='utf-8') as f:
                yaml.dump([finger], f, allow_unicode=True,
                         default_flow_style=False, sort_keys=False)
            success += 1
            print(f"  ✓ [{idx}/{total}] {finger_name} -> {output_file.relative_to(input_path.parent)}")
        except Exception as e:
            print(f"  ✗ [{idx}/{total}] {finger_name} - 写入失败: {e}")

    print(f"  完成: {success}/{total} 个指纹已拆分\n")


def main():
    """
    主函数
    """
    # 定义要处理的目录
    base_dir = Path(__file__).parent / "fingers"

    if not base_dir.exists():
        print(f"错误: fingers 目录不存在: {base_dir}")
        return

    print("=" * 60)
    print("指纹 YAML 文件拆分工具")
    print("=" * 60)
    print()

    # 遍历所有 YAML 文件
    yaml_files = list(base_dir.rglob("*.yaml")) + list(base_dir.rglob("*.yml"))

    if not yaml_files:
        print("未找到任何 YAML 文件")
        return

    print(f"找到 {len(yaml_files)} 个 YAML 文件\n")

    total_files = 0
    for yaml_file in yaml_files:
        # 拆分文件
        split_yaml_file(yaml_file)
        total_files += 1

    print("=" * 60)
    print(f"处理完成! 共处理 {total_files} 个文件")
    print(f"示例: cdn.yaml -> cdn/aliyun.yaml, cdn/cloudflare.yaml ...")
    print("=" * 60)


if __name__ == "__main__":
    main()
