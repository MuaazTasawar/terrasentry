from app.services.terraform_parser import parse_plan, build_plan_summary


def test_parse_plan_ignores_noop_changes():
    plan = {
        "resource_changes": [
            {"address": "aws_s3_bucket.unchanged", "type": "aws_s3_bucket",
             "provider_name": "aws", "change": {"actions": ["no-op"]}},
        ]
    }
    result = parse_plan(plan)
    assert result == []


def test_parse_plan_detects_create():
    plan = {
        "resource_changes": [
            {"address": "aws_instance.new", "type": "aws_instance",
             "provider_name": "aws", "change": {"actions": ["create"]}},
        ]
    }
    result = parse_plan(plan)
    assert len(result) == 1
    assert result[0].action == "create"
    assert result[0].address == "aws_instance.new"


def test_parse_plan_detects_replace_from_delete_create():
    plan = {
        "resource_changes": [
            {"address": "aws_instance.replaced", "type": "aws_instance",
             "provider_name": "aws", "change": {"actions": ["delete", "create"]}},
        ]
    }
    result = parse_plan(plan)
    assert len(result) == 1
    assert result[0].action == "replace"


def test_parse_plan_detects_delete():
    plan = {
        "resource_changes": [
            {"address": "aws_s3_bucket.old", "type": "aws_s3_bucket",
             "provider_name": "aws", "change": {"actions": ["delete"]}},
        ]
    }
    result = parse_plan(plan)
    assert result[0].action == "delete"


def test_build_plan_summary_empty_changes():
    summary = build_plan_summary([])
    assert "No resource changes" in summary


def test_build_plan_summary_lists_each_change():
    plan = {
        "resource_changes": [
            {"address": "aws_instance.new", "type": "aws_instance",
             "provider_name": "aws", "change": {"actions": ["create"]}},
            {"address": "aws_s3_bucket.old", "type": "aws_s3_bucket",
             "provider_name": "aws", "change": {"actions": ["delete"]}},
        ]
    }
    changes = parse_plan(plan)
    summary = build_plan_summary(changes)
    assert "2 resource change(s)" in summary
    assert "aws_instance.new" in summary
    assert "aws_s3_bucket.old" in summary