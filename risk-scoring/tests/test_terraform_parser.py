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


def test_parse_plan_extracts_tags_from_after_on_create():
    plan = {
        "resource_changes": [
            {"address": "aws_instance.new", "type": "aws_instance", "provider_name": "aws",
             "change": {"actions": ["create"], "after": {"tags": {"env": "prod"}}}},
        ]
    }
    result = parse_plan(plan)
    assert result[0].tags == {"env": "prod"}


def test_parse_plan_extracts_tags_from_before_on_delete():
    # A delete's `after` block is null, so tags must come from `before` —
    # this is what lets a policy rule catch "delete a prod-tagged resource".
    plan = {
        "resource_changes": [
            {"address": "aws_s3_bucket.old", "type": "aws_s3_bucket", "provider_name": "aws",
             "change": {"actions": ["delete"], "before": {"tags": {"env": "prod"}}, "after": None}},
        ]
    }
    result = parse_plan(plan)
    assert result[0].tags == {"env": "prod"}


def test_parse_plan_missing_tags_defaults_to_empty_dict():
    plan = {
        "resource_changes": [
            {"address": "aws_instance.untagged", "type": "aws_instance", "provider_name": "aws",
             "change": {"actions": ["create"], "after": {}}},
        ]
    }
    result = parse_plan(plan)
    assert result[0].tags == {}