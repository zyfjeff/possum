use super::*;
use std::borrow::Borrow;

#[derive(Clone)]
pub(crate) struct Dir {
    path_buf: PathBuf,
    block_size: u64,
}

impl AsRef<Path> for Dir {
    fn as_ref(&self) -> &Path {
        &self.path_buf
    }
}

impl Borrow<Path> for Dir {
    fn borrow(&self) -> &Path {
        &self.path_buf
    }
}

impl Dir {
    pub fn new(path_buf: PathBuf) -> Result<Self> {
        fs::create_dir_all(&path_buf)?;
        Ok(Self {
            path_buf,
            block_size: 4096,
        })
    }

    pub fn path(&self) -> &Path {
        &self.path_buf
    }

    pub fn block_size(&self) -> u64 {
        self.block_size
    }
}