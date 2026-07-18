import 'package:flutter/material.dart';
import '../models/approval_request.dart';
import '../services/api_service.dart';

class ApprovalDetailScreen extends StatefulWidget {
  final ApprovalRequest request;

  const ApprovalDetailScreen({super.key, required this.request});

  @override
  State<ApprovalDetailScreen> createState() => _ApprovalDetailScreenState();
}

class _ApprovalDetailScreenState extends State<ApprovalDetailScreen> {
  final ApiService _api = ApiService();
  bool _submitting = false;

  Future<void> _decide(String decision) async {
    setState(() => _submitting = true);
    try {
      await _api.decide(widget.request.id, decision);
      if (mounted) Navigator.pop(context);
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Failed to submit decision: $e')),
        );
      }
    } finally {
      if (mounted) setState(() => _submitting = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final r = widget.request;

    return Scaffold(
      appBar: AppBar(title: Text(r.repoName)),
      body: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Chip(
              label: Text('${r.riskLevel.toUpperCase()} · score ${r.riskScore}'),
            ),
            const SizedBox(height: 16),
            const Text('Reasoning', style: TextStyle(fontWeight: FontWeight.bold)),
            const SizedBox(height: 4),
            Text(r.reasoning.isNotEmpty ? r.reasoning : 'No reasoning provided.'),
            const SizedBox(height: 16),
            const Text('Plan Summary', style: TextStyle(fontWeight: FontWeight.bold)),
            const SizedBox(height: 4),
            Expanded(
              child: SingleChildScrollView(
                child: Text(r.planSummary, style: const TextStyle(fontFamily: 'monospace', fontSize: 13)),
              ),
            ),
            const SizedBox(height: 16),
            Row(
              children: [
                Expanded(
                  child: OutlinedButton(
                    onPressed: _submitting ? null : () => _decide('rejected'),
                    style: OutlinedButton.styleFrom(foregroundColor: Colors.red),
                    child: const Text('Reject'),
                  ),
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: FilledButton(
                    onPressed: _submitting ? null : () => _decide('approved'),
                    child: _submitting
                        ? const SizedBox(height: 18, width: 18, child: CircularProgressIndicator(strokeWidth: 2))
                        : const Text('Approve'),
                  ),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }
}