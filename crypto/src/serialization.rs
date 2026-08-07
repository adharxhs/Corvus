use crate::errors::{CryptoError, Result};
use serde::{Deserialize, Serialize};

pub fn serialize_bincode<T: Serialize>(val: &T) -> Result<Vec<u8>> {
    bincode::serialize(val).map_err(|e| CryptoError::Serialization(e.to_string()))
}

pub fn deserialize_bincode<'a, T: Deserialize<'a>>(bytes: &'a [u8]) -> Result<T> {
    bincode::deserialize(bytes).map_err(|e| CryptoError::Serialization(e.to_string()))
}

const BASE64_ALPHABET: &[u8; 64] =
    b"ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/";

pub fn encode_base64(data: &[u8]) -> String {
    let mut result = String::with_capacity(data.len().div_ceil(3) * 4);
    let mut i = 0;
    while i < data.len() {
        let b0 = data[i];
        let b1 = if i + 1 < data.len() { data[i + 1] } else { 0 };
        let b2 = if i + 2 < data.len() { data[i + 2] } else { 0 };

        let triplet = ((b0 as u32) << 16) | ((b1 as u32) << 8) | (b2 as u32);

        result.push(BASE64_ALPHABET[((triplet >> 18) & 0x3F) as usize] as char);
        result.push(BASE64_ALPHABET[((triplet >> 12) & 0x3F) as usize] as char);

        if i + 1 < data.len() {
            result.push(BASE64_ALPHABET[((triplet >> 6) & 0x3F) as usize] as char);
        } else {
            result.push('=');
        }

        if i + 2 < data.len() {
            result.push(BASE64_ALPHABET[(triplet & 0x3F) as usize] as char);
        } else {
            result.push('=');
        }

        i += 3;
    }
    result
}

pub fn decode_base64(input: &str) -> Result<Vec<u8>> {
    let clean: String = input.chars().filter(|c| !c.is_whitespace()).collect();
    if !clean.len().is_multiple_of(4) {
        return Err(CryptoError::Serialization(
            "Invalid base64 length".to_string(),
        ));
    }

    fn char_to_val(c: char) -> Result<u32> {
        match c {
            'A'..='Z' => Ok((c as u32) - ('A' as u32)),
            'a'..='z' => Ok((c as u32) - ('a' as u32) + 26),
            '0'..='9' => Ok((c as u32) - ('0' as u32) + 52),
            '+' => Ok(62),
            '/' => Ok(63),
            '=' => Ok(0),
            _ => Err(CryptoError::Serialization(format!(
                "Invalid base64 char: {}",
                c
            ))),
        }
    }

    let bytes = clean.as_bytes();
    let mut out = Vec::with_capacity(clean.len() / 4 * 3);
    let mut i = 0;
    while i < bytes.len() {
        let v0 = char_to_val(bytes[i] as char)?;
        let v1 = char_to_val(bytes[i + 1] as char)?;
        let v2 = char_to_val(bytes[i + 2] as char)?;
        let v3 = char_to_val(bytes[i + 3] as char)?;

        let triplet = (v0 << 18) | (v1 << 12) | (v2 << 6) | v3;

        out.push(((triplet >> 16) & 0xFF) as u8);
        if bytes[i + 2] != b'=' {
            out.push(((triplet >> 8) & 0xFF) as u8);
        }
        if bytes[i + 3] != b'=' {
            out.push((triplet & 0xFF) as u8);
        }
        i += 4;
    }
    Ok(out)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn base64_roundtrip() {
        let data = b"Hello Corvus Crypto!";
        let encoded = encode_base64(data);
        let decoded = decode_base64(&encoded).unwrap();
        assert_eq!(data.to_vec(), decoded);
    }
}
