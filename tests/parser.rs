use mdview::{model::Block, parser::parse_document};

#[test]
fn parses_core_blocks_into_owned_model() {
    let source = include_str!("../fixtures/basic.md");
    let document = parse_document(source, Some("basic.md".to_string()));

    assert!(matches!(
        document.blocks[0],
        Block::Heading { level: 1, .. }
    ));
    assert!(
        document
            .blocks
            .iter()
            .any(|block| matches!(block, Block::Callout { .. })),
        "expected callout block"
    );
    assert!(
        document
            .blocks
            .iter()
            .any(|block| matches!(block, Block::Table { .. })),
        "expected table block"
    );
    assert!(
        document
            .blocks
            .iter()
            .any(|block| matches!(block, Block::CodeBlock { .. })),
        "expected code block"
    );
    assert_eq!(document.headings.len(), 2);
}

#[test]
fn parses_task_list_state() {
    let document = parse_document("- [x] Done\n- [ ] Todo\n", None);
    let Block::List { items, .. } = &document.blocks[0] else {
        panic!("expected list");
    };

    assert_eq!(items[0].checked, Some(true));
    assert_eq!(items[1].checked, Some(false));
}
