use std::iter;

use puyo2::parse_ips_nazo_url;

#[test]
fn parse_ips_nazo_url_examples() {
    let cases = [
        (
            "https://ips.karou.jp/simu/pn.html?800F08J08A0EB_8161__270",
            "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaabaaaaafbaabaffaabaddaafadf",
            "pryr",
            "色ぷよ全て消す",
            [2, 7, 0],
        ),
        (
            "http://ips.karou.jp/simu/pn.html?jjgqqqqqqqqq_q1q1q1__u06",
            "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaececeacecececececececece",
            "gbgbgb",
            "6連鎖する",
            [30, 0, 6],
        ),
        (
            "http://ips.karou.jp/simu/pn.html?80080080oM0oM098_4141__u03",
            "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaabaaaaabaaaaabaaacagaaacagaaabbba",
            "brbr",
            "3連鎖する",
            [30, 0, 3],
        ),
    ];

    for (input, field, haipuyo, condition, cond_code) in cases {
        let decoded = parse_ips_nazo_url(input).unwrap();
        assert_eq!(decoded.initial_field, field);
        assert_eq!(decoded.haipuyo, haipuyo);
        assert_eq!(decoded.condition.text, condition);
        assert_eq!(decoded.condition_code, cond_code);
    }
}

#[test]
fn parse_ips_nazo_url_accepts_raw_query_and_path_style_input() {
    let from_url =
        parse_ips_nazo_url("https://ips.karou.jp/simu/pn.html?800F08J08A0EB_8161__270").unwrap();
    let from_raw = parse_ips_nazo_url("800F08J08A0EB_8161__270").unwrap();
    let from_path = parse_ips_nazo_url("pn.html?800F08J08A0EB_8161__270").unwrap();

    assert_eq!(from_url, from_raw);
    assert_eq!(from_url, from_path);
}

#[test]
fn parse_ips_nazo_url_preserves_internal_question_mark() {
    let with_question = parse_ips_nazo_url("800F08J08A0E?B_8161__270").unwrap();
    assert_eq!(
        with_question.initial_field,
        "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaabaaaaafbaabaffaabaddaafaaadf"
    );

    let truncated = parse_ips_nazo_url("B_8161__270").unwrap();
    assert_ne!(with_question.initial_field, truncated.initial_field);
}

#[test]
fn parse_ips_nazo_url_supports_data_mode() {
    let input = iter::once('~')
        .chain(iter::repeat_n('0', 77))
        .chain(iter::once('1'))
        .collect::<String>();
    let decoded = parse_ips_nazo_url(&input).unwrap();
    let want = iter::repeat_n('a', 77)
        .chain(iter::once('b'))
        .collect::<String>();
    assert_eq!(decoded.initial_field, want);
    assert!(decoded.haipuyo.is_empty());
}

#[test]
fn parse_ips_nazo_url_rejects_invalid_inputs() {
    assert!(parse_ips_nazo_url("!").is_err());
    assert!(parse_ips_nazo_url(".").is_err());
    assert!(parse_ips_nazo_url("?").is_err());
}
