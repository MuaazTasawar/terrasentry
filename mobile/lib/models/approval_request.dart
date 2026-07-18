class ApprovalRequest {
  final String id;
  final String repoName;
  final String planSummary;
  final int riskScore;
  final String riskLevel;
  final String reasoning;
  final String status;
  final DateTime createdAt;

  ApprovalRequest({
    required this.id,
    required this.repoName,
    required this.planSummary,
    required this.riskScore,
    required this.riskLevel,
    required this.reasoning,
    required this.status,
    required this.createdAt,
  });

  factory ApprovalRequest.fromJson(Map<String, dynamic> json) {
    return ApprovalRequest(
      id: json['id'] as String,
      repoName: json['repo_name'] as String,
      planSummary: json['plan_summary'] as String,
      riskScore: json['risk_score'] as int,
      riskLevel: json['risk_level'] as String,
      reasoning: json['reasoning'] as String? ?? '',
      status: json['status'] as String,
      createdAt: DateTime.tryParse(json['created_at'] as String? ?? '') ?? DateTime.now(),
    );
  }
}