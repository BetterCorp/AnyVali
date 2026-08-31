use anyvali::*;
use serde_json::json;

fn schema() -> ObjectSchema {
    object().field("id", Box::new(int())).field(
        "secret",
        Box::new(string().min_length(3).describe(
            "secret",
            Some(&DescribeOpts {
                sensitive: Some(true),
                ..Default::default()
            }),
        )),
    )
}

#[test]
fn encrypts_validates_and_decrypts_sensitive_fields() {
    let schema = schema();
    let input = json!({"id": 1, "secret": "clear"});
    let encrypted = encrypt(&schema, &input, |path, value| {
        assert_eq!(path, &[PathSegment::Key("secret".to_string())]);
        Ok::<_, String>(json!(format!("encrypted:{}", value.as_str().unwrap())))
    })
    .unwrap();

    assert_eq!(encrypted, json!({"id": 1, "secret": "encrypted:clear"}));
    assert!(safe_parse_encrypted(&schema, &encrypted).success);
    assert!(!safe_parse_encrypted(&schema, &json!({"id": 1, "secret": "clear"})).success);

    let decrypted = decrypt(&schema, &encrypted, |_path, value| {
        Ok::<_, String>(json!(value
            .as_str()
            .unwrap()
            .trim_start_matches("encrypted:")))
    })
    .unwrap();
    assert_eq!(decrypted, input);
}

#[test]
fn sensitive_objects_are_opaque_and_bad_callbacks_are_rejected() {
    let credentials = object().field("user", Box::new(string())).describe(
        "credentials",
        Some(&DescribeOpts {
            sensitive: Some(true),
            ..Default::default()
        }),
    );
    let schema = object().field("credentials", Box::new(credentials));
    let input = json!({"credentials": {"user": "alice"}});

    let result = encrypt(&schema, &input, |_path, _value| {
        Ok::<_, String>(json!("broken"))
    });
    assert!(matches!(result, Err(SensitiveError::Validation(_))));
}
