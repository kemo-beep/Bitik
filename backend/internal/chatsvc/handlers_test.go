package chatsvc

import "testing"

func TestPageMetaHasNext(t *testing.T) {
	meta := pageMeta(1, 20, 21)
	if meta["has_next"] != true {
		t.Fatalf("expected has_next true, got %#v", meta["has_next"])
	}
	meta2 := pageMeta(1, 20, 20)
	if meta2["has_next"] != false {
		t.Fatalf("expected has_next false, got %#v", meta2["has_next"])
	}
}

