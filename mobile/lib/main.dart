import 'package:flutter/material.dart';

void main() {
  runApp(const TerraSentryApp());
}

class TerraSentryApp extends StatelessWidget {
  const TerraSentryApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'TerraSentry',
      debugShowCheckedModeBanner: false,
      theme: ThemeData(
        colorSchemeSeed: const Color(0xFF2E7D32),
        useMaterial3: true,
        brightness: Brightness.dark,
      ),
      home: const _PlaceholderHome(),
    );
  }
}

// Replaced by HomeScreen in Phase 4 (mobile/lib/screens/home_screen.dart).
class _PlaceholderHome extends StatelessWidget {
  const _PlaceholderHome();

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('TerraSentry')),
      body: const Center(
        child: Text('On-call approvals coming in Phase 4'),
      ),
    );
  }
}