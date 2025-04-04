import type React from 'react';
import { useEffect, useState } from 'react';
import { Container, TextField, MenuItem, Button, Typography, Box, Stack } from '@mui/material';
import { DownloadYoutube } from "../wailsjs/go/main/App";
import { EventsEmit, EventsOn } from '../wailsjs/runtime';

const YTDownloadPage: React.FC = () => {
    const [url, setUrl] = useState('');
    const [format, setFormat] = useState('video');
    const [quality, setQuality] = useState('best');
    const [log, setLog] = useState('');
    const [isDownloading, setIsDownloading] = useState(false);

    const handleDownload = () => {
        setIsDownloading(true);
        setLog('Starting download...\n');
        DownloadYoutube(url, format, quality).then(
            () => {
                setLog((prevLog) => `${prevLog}Download completed\n`);
                setIsDownloading(false);
            }
        ).catch((error) => {
            setLog((prevLog) => `${prevLog}Error: ${error.message}\n`);
            setIsDownloading(false);
        });
    };

    useEffect(() => {
        // Listen for download progress events
        EventsOn("download-progress", (message: string) => {
            setLog((prevLog) => `${prevLog}${message}`);
        });

        // Cleanup event listener on component unmount
        return () => {
            EventsOn("download-progress", () => { }); // Unsubscribe
        };
    }, []);

    return (
        <Container maxWidth="sm" style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', minHeight: 'calc(100% - navbarpx)' }}>
            <Typography variant="h4" gutterBottom color='black'>
                YouTube Downloader
            </Typography>
            <Box component="form" noValidate autoComplete="off" style={{ width: '100%' }}>
                <Stack>
                    <TextField
                        fullWidth
                        label="YouTube URL"
                        variant="outlined"
                        margin="normal"
                        value={url}
                        onChange={(e) => setUrl(e.target.value)}
                    />
                    <TextField
                        fullWidth
                        select
                        label="Format"
                        variant="outlined"
                        margin="normal"
                        value={format}
                        onChange={(e) => setFormat(e.target.value)}
                    >
                        <MenuItem value="video">Video</MenuItem>
                        <MenuItem value="audio">Audio</MenuItem>
                    </TextField>
                    {format === 'video' && (
                        <TextField
                            fullWidth
                            select
                            label="Quality"
                            variant="outlined"
                            margin="normal"
                            value={quality}
                            onChange={(e) => setQuality(e.target.value)}
                        >
                            <MenuItem value="best">BEST</MenuItem>
                            <MenuItem value="360">360p</MenuItem>
                            <MenuItem value="480">480p</MenuItem>
                            <MenuItem value="720">720p</MenuItem>
                            <MenuItem value="1080">1080p</MenuItem>
                            <MenuItem value="2160">4k</MenuItem>
                        </TextField>
                    )}
                    {isDownloading ? (
                        <Button
                            fullWidth
                            variant="contained"
                            color="secondary"
                            onClick={() => {
                                setIsDownloading(false);
                                EventsEmit("cancel-download", () => { });
                            }}
                            style={{ marginTop: '16px' }}
                        >
                            Cancel
                        </Button>
                    ) : (
                        <Button
                            fullWidth
                            variant="contained"
                            color="primary"
                            onClick={handleDownload}
                            style={{ marginTop: '16px' }}
                            disabled={isDownloading}
                        >
                            Download
                        </Button>
                    )}

                    <TextField
                        fullWidth
                        label="Console"
                        variant="outlined"
                        margin="normal"
                        multiline
                        rows={10}
                        value={log}
                        InputProps={{
                            readOnly: true,
                        }}
                        style={{ marginTop: '16px', backgroundColor: '#f5f5f5' }}
                        inputRef={(input) => {
                            if (input) {
                                input.scrollTop = input.scrollHeight;
                            }
                        }}
                    />
                </Stack>
            </Box>
        </Container >
    );
};

export default YTDownloadPage;