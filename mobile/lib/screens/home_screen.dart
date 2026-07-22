import 'package:flutter/material.dart';
import '../models/approval_request.dart';
import '../services/api_service.dart';
import '../widgets/approval_card.dart';
import 'approval_detail_screen.dart';
import 'login_screen.dart';

class HomeScreen extends StatefulWidget {
  const HomeScreen({super.key});

  @override
  State<HomeScreen> createState() => _HomeScreenState();
}

class _HomeScreenState extends State<HomeScreen> {
  final ApiService _api = ApiService();
  late Future<List<ApprovalRequest>> _pendingScans;

  @override
  void initState() {
    super.initState();
    _pendingScans = _api.fetchPendingScans();
  }

  Future<void> _refresh() async {
    setState(() {
      _pendingScans = _api.fetchPendingScans();
    });
    await _pendingScans;
  }

  Future<void> _logout() async {
    await _api.logout();
    if (mounted) {
      Navigator.of(context).pushAndRemoveUntil(
        MaterialPageRoute(builder: (_) => const LoginScreen()),
        (route) => false,
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Pending Approvals'),
        actions: [
          IconButton(
            icon: const Icon(Icons.logout),
            tooltip: 'Log out',
            onPressed: _logout,
          ),
        ],
      ),
      body: RefreshIndicator(
        onRefresh: _refresh,
        child: FutureBuilder<List<ApprovalRequest>>(
          future: _pendingScans,
          builder: (context, snapshot) {
            if (snapshot.connectionState == ConnectionState.waiting) {
              return const Center(child: CircularProgressIndicator());
            }
            if (snapshot.hasError) {
              return ListView(
                children: [
                  const SizedBox(height: 80),
                  Icon(Icons.cloud_off, size: 48, color: Colors.grey.shade600),
                  const SizedBox(height: 12),
                  Center(child: Text('Could not reach TerraSentry API\n${snapshot.error}',
                      textAlign: TextAlign.center)),
                ],
              );
            }

            final scans = snapshot.data ?? [];
            if (scans.isEmpty) {
              return ListView(
                children: const [
                  SizedBox(height: 120),
                  Icon(Icons.check_circle_outline, size: 48, color: Colors.green),
                  SizedBox(height: 12),
                  Center(child: Text('No pending approvals — all clear')),
                ],
              );
            }

            return ListView.builder(
              itemCount: scans.length,
              itemBuilder: (context, index) {
                final scan = scans[index];
                return ApprovalCard(
                  request: scan,
                  onTap: () async {
                    await Navigator.push(
                      context,
                      MaterialPageRoute(builder: (_) => ApprovalDetailScreen(request: scan)),
                    );
                    _refresh();
                  },
                );
              },
            );
          },
        ),
      ),
    );
  }
}