#!/usr/bin/env python3
"""
Extract enum values from botocore service models.

This script extracts enum types from botocore to provide a single source
of truth for AWS enum values used by both Python and Go code generators.

Usage:
    python scripts/extract_enums.py [--output enums/enums.json]

Requirements:
    pip install botocore
"""

import argparse
import json
import re
import sys
from collections import defaultdict
from pathlib import Path


def to_go_const_name(service: str, enum_name: str, value: str) -> str:
    """
    Convert an enum value to a valid Go constant name with service prefix.

    Examples:
        ("lambda", "Runtime", "python3.12") -> "LambdaRuntimePython312"
        ("s3", "StorageClass", "STANDARD_IA") -> "S3StorageClassStandardIa"
        ("ec2", "InstanceType", "t2.micro") -> "Ec2InstanceTypeT2Micro"
    """
    # Start with capitalized service name
    result = capitalize_service(service) + enum_name

    # Normalize value: replace all non-alphanumeric with spaces for word splitting
    normalized = re.sub(r"[^a-zA-Z0-9]+", " ", value)

    # Capitalize each word
    for word in normalized.split():
        if word:
            # Capitalize first letter, lowercase rest
            result += word[0].upper() + word[1:].lower()

    return result


def capitalize_service(service: str) -> str:
    """Capitalize service name for Go constant prefix."""
    special_cases = {
        "ec2": "Ec2",
        "ecs": "Ecs",
        "rds": "Rds",
        "s3": "S3",
        "acm": "Acm",
        "elbv2": "Elbv2",
    }
    return special_cases.get(service, service.capitalize())


def to_python_const_name(value: str) -> str:
    """
    Convert an enum value to a valid Python constant name.

    Examples:
        "python3.12" -> "PYTHON3_12"
        "PAY_PER_REQUEST" -> "PAY_PER_REQUEST"
        "t2.micro" -> "T2_MICRO"
    """
    # Replace dots and hyphens with underscores
    name = value.replace(".", "_").replace("-", "_")
    # Remove any other non-alphanumeric characters
    name = re.sub(r"[^a-zA-Z0-9_]", "", name)
    # Convert to uppercase
    name = name.upper()
    # Ensure it starts with a letter or underscore
    if name and name[0].isdigit():
        name = "_" + name
    return name


def extract_all_enums() -> dict[str, dict[str, list[str]]]:
    """
    Extract all enums from botocore for all available services.

    Returns:
        Dict mapping service -> shape_name -> list of enum values
    """
    try:
        import botocore.loaders
        import botocore.session
    except ImportError:
        print("ERROR: botocore not installed. Run: pip install botocore", file=sys.stderr)
        sys.exit(1)

    loader = botocore.loaders.Loader()
    session = botocore.session.get_session()

    # Get list of available services
    available_services = session.get_available_services()

    result: dict[str, dict[str, list[str]]] = defaultdict(dict)

    for service_name in available_services:
        try:
            service_model = loader.load_service_model(service_name, "service-2")
            shapes = service_model.get("shapes", {})

            for shape_name, shape_def in shapes.items():
                if shape_def.get("type") == "string" and "enum" in shape_def:
                    enum_values = shape_def["enum"]
                    # Filter out very large enums (likely not useful as constants)
                    if len(enum_values) <= 500:
                        result[service_name][shape_name] = enum_values

        except Exception:
            # Service might not have service-2 model
            continue

    return dict(result)


def generate_enum_data(all_enums: dict[str, dict[str, list[str]]]) -> dict:
    """
    Generate the enum data structure for JSON output.

    Returns dict with structure:
    {
        "services": {
            "lambda": {
                "Runtime": {
                    "name": "Runtime",
                    "values": [
                        {"value": "python3.12", "goName": "LambdaRuntimePython312", "pyName": "PYTHON3_12"},
                        ...
                    ]
                }
            }
        }
    }
    """
    output = {"services": {}}

    for service, shapes in sorted(all_enums.items()):
        service_enums = {}
        for shape_name, values in sorted(shapes.items()):
            enum_values = []
            for v in values:
                go_name = to_go_const_name(service, shape_name, v)
                py_name = to_python_const_name(v)
                if py_name:  # Filter out empty names
                    enum_values.append({
                        "value": v,
                        "goName": go_name,
                        "pyName": py_name,
                    })
            if enum_values:
                service_enums[shape_name] = {
                    "name": shape_name,
                    "values": enum_values,
                }
        if service_enums:
            output["services"][service] = service_enums

    return output


def main():
    parser = argparse.ArgumentParser(
        description="Extract enum values from botocore service models"
    )
    parser.add_argument(
        "--output",
        type=Path,
        default=Path(__file__).parent.parent / "enums" / "enums.json",
        help="Output file path (default: enums/enums.json)",
    )
    args = parser.parse_args()

    print("Extracting enums from botocore...")

    all_enums = extract_all_enums()

    # Count totals
    total_services = len(all_enums)
    total_enums = sum(len(shapes) for shapes in all_enums.values())
    total_values = sum(
        len(values) for shapes in all_enums.values() for values in shapes.values()
    )

    print(f"  Found {total_enums} enum types across {total_services} services")
    print(f"  Total {total_values} enum values")

    # Generate output
    output = generate_enum_data(all_enums)

    # Ensure output directory exists
    args.output.parent.mkdir(parents=True, exist_ok=True)

    # Write to file
    args.output.write_text(json.dumps(output, indent=2, sort_keys=True))
    print(f"\nEnums written to {args.output}")
    print("Enum extraction completed successfully!")


if __name__ == "__main__":
    main()
