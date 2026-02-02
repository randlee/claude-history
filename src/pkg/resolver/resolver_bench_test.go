package resolver

import (
	"testing"
	"time"
)

// BenchmarkResolveSessionID_Small benchmarks prefix resolution with 10 sessions.
func BenchmarkResolveSessionID_Small(b *testing.B) {
	projectDir := createLargeTestProject(b, 10)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ResolveSessionID(projectDir, "00000001")
	}
}

// BenchmarkResolveSessionID_Medium benchmarks prefix resolution with 50 sessions.
func BenchmarkResolveSessionID_Medium(b *testing.B) {
	projectDir := createLargeTestProject(b, 50)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ResolveSessionID(projectDir, "00000001")
	}
}

// BenchmarkResolveSessionID_Large benchmarks prefix resolution with 100 sessions.
func BenchmarkResolveSessionID_Large(b *testing.B) {
	projectDir := createLargeTestProject(b, 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ResolveSessionID(projectDir, "00000002")
	}
}

// BenchmarkResolveSessionID_VeryLarge benchmarks prefix resolution with 500 sessions.
func BenchmarkResolveSessionID_VeryLarge(b *testing.B) {
	projectDir := createLargeTestProject(b, 500)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ResolveSessionID(projectDir, "00000003")
	}
}

// BenchmarkResolveSessionID_ShortPrefix benchmarks with short prefix (more comparisons).
func BenchmarkResolveSessionID_ShortPrefix(b *testing.B) {
	projectDir := createLargeTestProject(b, 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ResolveSessionID(projectDir, "000")
	}
}

// BenchmarkResolveSessionID_LongPrefix benchmarks with long prefix (faster resolution).
func BenchmarkResolveSessionID_LongPrefix(b *testing.B) {
	projectDir := createLargeTestProject(b, 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ResolveSessionID(projectDir, "00000001")
	}
}

// BenchmarkResolveAgentID benchmarks agent ID resolution.
func BenchmarkResolveAgentID(b *testing.B) {
	projectDir, sessions := createTestProjectStructure(&testing.T{})

	// Extract session ID from first session
	sessionID := "aaa12345-1234-1234-1234-123456789abc"
	sessionDir := projectDir + "/" + sessionID

	// Create multiple agents
	for i := 0; i < 10; i++ {
		agentID := generateRandomUUID(i)
		createTestAgent(&testing.T{}, sessionDir, agentID, "Agent "+string(rune('A'+i)))
	}

	_ = sessions // Use the variable to avoid compiler error

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ResolveAgentID(projectDir, sessionID, "0")
	}
}

// BenchmarkResolveSessionID_Parallel benchmarks concurrent prefix resolution.
func BenchmarkResolveSessionID_Parallel(b *testing.B) {
	projectDir := createLargeTestProject(b, 100)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = ResolveSessionID(projectDir, "00000001")
		}
	})
}

// TestResolveSessionID_LargeProject tests with 100+ sessions.
func TestResolveSessionID_LargeProject(t *testing.T) {
	projectDir := createLargeTestProject(t, 150)

	// Test various prefixes that should be unique with our UUID generation
	// UUIDs start with 8-digit hex of index: 00000000, 00000001, 00000002, etc.
	tests := []struct {
		prefix  string
		wantErr bool
	}{
		{"00000000", false}, // Index 0
		{"00000001", false}, // Index 1
		{"0000000a", false}, // Index 10
		{"zzz", true},       // Should not exist
	}

	for _, tt := range tests {
		t.Run("prefix_"+tt.prefix, func(t *testing.T) {
			_, err := ResolveSessionID(projectDir, tt.prefix)
			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// TestResolveAgentID_DeeplyNested tests agent resolution with deep nesting.
func TestResolveAgentID_DeeplyNested(t *testing.T) {
	projectDir, _ := createTestProjectStructure(t)

	sessionID := "deep1234-1234-1234-1234-123456789abc"
	createTestSession(t, projectDir, sessionID, "Deep test", now())

	// Create deeply nested agent structure (5 levels)
	createDeeplyNestedAgents(t, projectDir, sessionID, 5)

	// Should be able to resolve agents at various depths
	// Note: The actual test depends on how deep agent discovery works
	// For now, just verify we can access the structure

	// This is a placeholder - actual implementation depends on resolver logic
	t.Log("Created deeply nested agent structure with 5 levels")
}

// BenchmarkResolveSessionID_Ambiguous benchmarks ambiguous prefix error generation.
func BenchmarkResolveSessionID_Ambiguous(b *testing.B) {
	projectDir := b.TempDir()

	// Create many sessions with same prefix "ambig"
	for i := 0; i < 10; i++ {
		sessionID := "ambig" + generateUniqueUUIDSuffix(i)[0:3] + "-1111-1111-1111-111111111111"
		createTestSessionDirect(b, projectDir, sessionID, "Ambiguous test", now().Add(time.Duration(-i)*time.Hour))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ResolveSessionID(projectDir, "ambig")
	}
}

// TestResolveSessionID_PerformanceRequirement tests that resolution is fast enough.
func TestResolveSessionID_PerformanceRequirement(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping performance test in short mode")
	}

	projectDir := createLargeTestProject(t, 100)

	// Measure resolution time with a unique prefix
	start := now()
	_, err := ResolveSessionID(projectDir, "00000001") // Unique prefix for index 1
	duration := now().Sub(start)

	if err != nil {
		t.Fatalf("resolution failed: %v", err)
	}

	// Should complete in under 10ms for 100 sessions
	maxDuration := 10_000_000 // 10ms in nanoseconds
	if duration.Nanoseconds() > int64(maxDuration) {
		t.Errorf("resolution took %v, want < 10ms", duration)
	}

	t.Logf("Resolution took %v for 100 sessions", duration)
}

// now returns current time.
func now() time.Time {
	return time.Now()
}
