import 'dart:convert';
import 'package:http/http.dart' as http;
import 'package:shared_preferences/shared_preferences.dart';
import '../models/approval_request.dart';

/// Thrown when the API rejects a request with 401 — the stored token is
/// missing, invalid, or expired. Screens catch this specifically to send
/// the user back to the login screen instead of showing a generic error.
class UnauthorizedException implements Exception {
  final String message;
  UnauthorizedException([this.message = 'Session expired — please sign in again.']);

  @override
  String toString() => message;
}

class ApiService {
  // Points at the local Go API by default. For a physical device on the
  // same network, swap to your machine's LAN IP; for an Android emulator,
  // use 10.0.2.2 instead of localhost.
  final String baseUrl;

  ApiService({this.baseUrl = 'http://localhost:8080'});

  static const _tokenKey = 'terrasentry_auth_token';

  Future<String?> _getToken() async {
    final prefs = await SharedPreferences.getInstance();
    return prefs.getString(_tokenKey);
  }

  Future<void> _saveToken(String token) async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(_tokenKey, token);
  }

  Future<bool> isLoggedIn() async => (await _getToken()) != null;

  Future<void> logout() async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.remove(_tokenKey);
  }

  Future<Map<String, String>> _authHeaders() async {
    final token = await _getToken();
    return {
      'Content-Type': 'application/json',
      if (token != null) 'Authorization': 'Bearer $token',
    };
  }

  /// Logs in against the Go API and persists the returned JWT on-device.
  Future<void> login(String email, String password) async {
    final response = await http.post(
      Uri.parse('$baseUrl/api/v1/auth/login'),
      headers: {'Content-Type': 'application/json'},
      body: jsonEncode({'email': email, 'password': password}),
    );

    if (response.statusCode != 200) {
      throw Exception('Invalid email or password');
    }

    final body = jsonDecode(response.body) as Map<String, dynamic>;
    await _saveToken(body['token'] as String);
  }

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
      headers: await _authHeaders(),
      body: jsonEncode({'decision': decision, 'decided_by': decidedBy}),
    );

    if (response.statusCode == 401) {
      throw UnauthorizedException();
    }
    if (response.statusCode != 200) {
      throw Exception('Failed to record decision (${response.statusCode})');
    }
  }

  Future<void> registerDevice(String deviceToken, {String ownerName = 'on-call'}) async {
    final response = await http.post(
      Uri.parse('$baseUrl/api/v1/devices'),
      headers: await _authHeaders(),
      body: jsonEncode({'device_token': deviceToken, 'owner_name': ownerName}),
    );

    if (response.statusCode == 401) {
      throw UnauthorizedException();
    }
  }
}