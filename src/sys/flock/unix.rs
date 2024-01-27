use super::*;
use libc::*;
use std::io::SeekFrom;

use nix::fcntl::FlockArg;
pub use nix::fcntl::FlockArg::*;

const EWOULDBLOCK: Errno = Errno::EWOULDBLOCK;

pub fn try_lock_file(file: &mut File, arg: FlockArg) -> nix::Result<bool> {
    let flock_res = if false {
        nix::fcntl::flock(file.as_raw_fd(), arg)
    } else {
        lock_file_segment(file, arg, None, SeekFrom::Start(0))
    };
    match flock_res {
        Ok(()) => Ok(true),
        Err(errno) => {
            if errno == EWOULDBLOCK {
                Ok(false)
            } else {
                Err(errno)
            }
        }
    }
}

fn seek_from_offset(seek_from: SeekFrom) -> off_t {
    use SeekFrom::*;
    match seek_from {
        Start(offset) => offset as off_t,
        End(offset) | Current(offset) => offset as off_t,
    }
}

fn seek_from_whence(seek_from: SeekFrom) -> c_short {
    use libc::*;
    use SeekFrom::*;
    (match seek_from {
        Start(_) => SEEK_SET,
        Current(_) => SEEK_CUR,
        End(_) => SEEK_END,
    }) as c_short
}

pub fn lock_file_segment(
    file: &File,
    arg: FlockArg,
    len: Option<i64>,
    whence: SeekFrom,
) -> nix::Result<()> {
    if let Some(len) = len {
        // This has special meaning on macOS: To the end of the file. Use None instead.
        assert_ne!(len, 0);
    }
    let flock_arg = libc::flock {
        l_start: seek_from_offset(whence),
        l_len: len.unwrap_or_default(),
        l_pid: 0,
        l_type: match arg {
            LockShared | LockSharedNonblock => libc::F_RDLCK,
            LockExclusive | LockExclusiveNonblock => libc::F_WRLCK,
            Unlock | UnlockNonblock => libc::F_UNLCK,
            // Silly non-exhaustive enum.
            _ => unimplemented!(),
        },
        l_whence: seek_from_whence(whence),
    };
    let cmd = match arg {
        LockShared | LockExclusive | Unlock => 91,
        LockSharedNonblock | LockExclusiveNonblock | UnlockNonblock => 90,
        _ => unimplemented!(),
    };
    // nix::fcntl::fcntl()
    let nix_result: nix::Result<c_int> = nix::errno::Errno::result(unsafe {
        libc::fcntl(file.as_raw_fd(), cmd, &flock_arg as *const _)
    });
    nix_result?;
    Ok(())
}
