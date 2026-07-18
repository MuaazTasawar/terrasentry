import 'dart:convert';
import 'package:http/http.dart' as http;
import '../models/approval_request.dart';

class ApiService {
  // Points at the local Go API by default. For a physical device on the
  // same network, swap to your machine's LAN IP; for an Android emulator,
  // use 10.0.2.2 instead of localhost.
  final String baseUrl;

  ApiService({this.baseUrl = 'http://localhost:8080'});

  Future<List<ApprovalRequest>> fetchPendingScans() async {
    final response = await http.get(Uri.parse('$baseUrl/api/v1/scans/pending'));

    if (response.statusCode != 200) {
      throw Exception('Failed to load pending scans (${response.statusCode})');
    }

    final body = jsonDecode(response.body) as Map<String, dynamic>;
    final scans = body['scans'] as List<dynamic>;
    return scans
        .map((s) => ApprovalRequest.fromJson(s as Map<String, dynamic>))
        .toList();
  }

  Future<void> decide(String scanId, String decision, {String decidedBy = 'on-call'}) async {
    final response = await http.post(
      Uri.parse('$baseUrl/api/v1/scans/$scanId/decision'),
      headers: {'Content-Type': 'application/json'},
      body: jsonEncode({'decision': decision, 'decided_by': decidedBy}),
    );

    if (response.statusCode != 200) {
      throw Exception('Failed to record decision (${response.statusCode})');
    }
  }

  Future<void> registerDevice(String deviceToken, {String ownerName = 'on-call'}) async {
    await http.post(
      Uri.parse('$baseUrl/api/v1/devices'),
      headers: {'Content-Type': 'application/json'},
      body: jsonEncode({'device_token': deviceToken, 'owner_name': ownerName}),
    );
  }
}