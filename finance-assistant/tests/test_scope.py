from finance_assistant.scope import ConversationScope, owner_workspace_path


def test_channel_threads_are_separate_but_memory_namespace_is_shared() -> None:
    app = ConversationScope("owner-a", "app", "native-default")
    telegram = ConversationScope("owner-a", "telegram", "123")
    other = ConversationScope("owner-b", "app", "native-default")

    assert app.thread_id != telegram.thread_id
    assert app.memory_namespace == telegram.memory_namespace
    assert app.memory_namespace != other.memory_namespace


def test_workspace_path_uses_opaque_hash_and_cannot_traverse(tmp_path) -> None:
    path = owner_workspace_path(tmp_path, "../../owner-a")
    assert path.parent == tmp_path / "workspaces"
    assert "owner-a" not in path.name
    assert ".." not in path.name
