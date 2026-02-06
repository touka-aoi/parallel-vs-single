package application

import (
	"context"
	"testing"

	"withered/server/domain"
)

func TestNewField(t *testing.T) {
	m := NewMap(10, 10, 1.0)
	f := NewField(m)

	if f.Map != m {
		t.Error("Map not set correctly")
	}
	if len(f.Actors) != 0 {
		t.Errorf("Actors length = %d, want 0", len(f.Actors))
	}
}

func TestField_SpawnAtCenter(t *testing.T) {
	m := NewMap(10, 10, 1.0) // WorldWidth=10, WorldHeight=10
	f := NewField(m)

	sessionID := domain.NewSessionID()
	actor := f.SpawnAtCenter(sessionID)

	if actor.SessionID != sessionID {
		t.Errorf("SessionID = %s, want %s", actor.SessionID, sessionID)
	}
	if actor.Position.X != 5.0 || actor.Position.Y != 5.0 {
		t.Errorf("Position = (%f, %f), want (5, 5)", actor.Position.X, actor.Position.Y)
	}
	if len(f.Actors) != 1 {
		t.Errorf("Actors length = %d, want 1", len(f.Actors))
	}
}

func TestField_Move(t *testing.T) {
	m := NewMap(10, 10, 1.0) // WorldWidth=10, WorldHeight=10
	f := NewField(m)
	ctx := context.Background()

	sessionID := domain.NewSessionID()
	f.SpawnAtCenter(sessionID) // (5, 5)

	f.ActorMove(ctx, sessionID, 2.0, -1.0)

	actor, ok := f.GetActor(sessionID)
	if !ok {
		t.Fatal("actor not found")
	}
	if actor.Position.X != 7.0 || actor.Position.Y != 4.0 {
		t.Errorf("Position = (%f, %f), want (7, 4)", actor.Position.X, actor.Position.Y)
	}
}

func TestField_Move_Clamp(t *testing.T) {
	m := NewMap(10, 10, 1.0) // WorldWidth=10, WorldHeight=10
	f := NewField(m)
	ctx := context.Background()

	sessionID := domain.NewSessionID()
	f.SpawnAtCenter(sessionID) // (5, 5)

	tests := []struct {
		name      string
		dx, dy    float32
		expectedX float32
		expectedY float32
	}{
		{"clamp max x", 100, 0, 10.0, 5.0},
		{"clamp min x", -100, 0, 0.0, 5.0},
		{"clamp max y", 0, 100, 0.0, 10.0},
		{"clamp min y", 0, -100, 0.0, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f.ActorMove(ctx, sessionID, tt.dx, tt.dy)
			actor, _ := f.GetActor(sessionID)
			if actor.Position.X != tt.expectedX || actor.Position.Y != tt.expectedY {
				t.Errorf("Position = (%f, %f), want (%f, %f)",
					actor.Position.X, actor.Position.Y, tt.expectedX, tt.expectedY)
			}
		})
	}
}

func TestField_Move_ActorNotFound(t *testing.T) {
	m := NewMap(10, 10, 1.0)
	f := NewField(m)
	ctx := context.Background()

	// 存在しないアクターへのMoveはパニックしない（警告ログのみ）
	nonExistentID := domain.NewSessionID()
	f.ActorMove(ctx, nonExistentID, 1.0, 1.0)
}

func TestField_Remove(t *testing.T) {
	m := NewMap(10, 10, 1.0)
	f := NewField(m)

	sessionID1 := domain.NewSessionID()
	sessionID2 := domain.NewSessionID()
	f.SpawnAtCenter(sessionID1)
	f.SpawnAtCenter(sessionID2)

	if len(f.Actors) != 2 {
		t.Fatalf("Actors length = %d, want 2", len(f.Actors))
	}

	f.Remove(sessionID1)

	if len(f.Actors) != 1 {
		t.Errorf("Actors length = %d, want 1", len(f.Actors))
	}
	if _, ok := f.GetActor(sessionID1); ok {
		t.Error("actor 1 should be removed")
	}
	if _, ok := f.GetActor(sessionID2); !ok {
		t.Error("actor 2 should exist")
	}
}

func TestField_GetAllActors(t *testing.T) {
	m := NewMap(10, 10, 1.0)
	f := NewField(m)

	f.SpawnAtCenter(domain.NewSessionID())
	f.SpawnAtCenter(domain.NewSessionID())
	f.SpawnAtCenter(domain.NewSessionID())

	actors := f.GetAllActors()
	if len(actors) != 3 {
		t.Errorf("actors length = %d, want 3", len(actors))
	}
}
