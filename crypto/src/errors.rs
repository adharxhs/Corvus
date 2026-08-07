use std::fmt;

#[derive(Debug)]
pub enum CryptoError {
    Identity(String),
    Prekey(String),
    Handshake(String),
    Ratchet(String),
    SenderKey(String),
    Serialization(String),
    Storage(String),
}

impl fmt::Display for CryptoError {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            CryptoError::Identity(msg) => write!(f, "Identity error: {}", msg),
            CryptoError::Prekey(msg) => write!(f, "Prekey error: {}", msg),
            CryptoError::Handshake(msg) => write!(f, "Handshake error: {}", msg),
            CryptoError::Ratchet(msg) => write!(f, "Ratchet error: {}", msg),
            CryptoError::SenderKey(msg) => write!(f, "SenderKey error: {}", msg),
            CryptoError::Serialization(msg) => write!(f, "Serialization error: {}", msg),
            CryptoError::Storage(msg) => write!(f, "Storage error: {}", msg),
        }
    }
}

impl std::error::Error for CryptoError {}

pub type Result<T> = std::result::Result<T, CryptoError>;
