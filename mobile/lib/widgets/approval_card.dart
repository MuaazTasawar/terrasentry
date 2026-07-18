import 'package:flutter/material.dart';
import '../models/approval_request.dart';

class ApprovalCard extends StatelessWidget {
  final ApprovalRequest request;
  final VoidCallback onTap;

  const ApprovalCard({super.key, required this.request, required this.onTap});

  Color _riskColor(BuildContext context) {
    switch (request.riskLevel) {
      case 'high':
        return Colors.red.shade400;
      case 'medium':
        return Colors.orange.shade400;
      default:
        return Colors.green.shade400;
    }
  }

  @override
  Widget build(BuildContext context) {
    return Card(
      margin: const EdgeInsets.symmetric(horizontal: 16, vertical: 6),
      child: ListTile(
        onTap: onTap,
        contentPadding: const EdgeInsets.all(12),
        leading: CircleAvatar(
          backgroundColor: _riskColor(context),
          child: Text(
            request.riskScore.toString(),
            style: const TextStyle(color: Colors.white, fontSize: 12, fontWeight: FontWeight.bold),
          ),
        ),
        title: Text(request.repoName, style: const TextStyle(fontWeight: FontWeight.w600)),
        subtitle: Text(
          request.reasoning.isNotEmpty ? request.reasoning : request.planSummary,
          maxLines: 2,
          overflow: TextOverflow.ellipsis,
        ),
        trailing: Chip(
          label: Text(request.riskLevel.toUpperCase(), style: const TextStyle(fontSize: 11)),
          backgroundColor: _riskColor(context).withOpacity(0.15),
        ),
      ),
    );
  }
}