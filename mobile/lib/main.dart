import 'package:flutter/material.dart';
import 'screens/home_screen.dart';

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
      home: const HomeScreen(),
    );
  }
}