import React, { useState } from 'react';
import { Container, TextField, MenuItem, Button, Typography, Box, Stack } from '@mui/material';
import { DownloadYoutube } from "../wailsjs/go/main/App";

const YTDownloadPage: React.FC = () => {
    const [url, setUrl] = useState('');
    const [format, setFormat] = useState('video');
    const [quality, setQuality] = useState('360p');

    const handleDownload = () => {
        // Implement download logic here
        console.log(`URL: ${url}, Format: ${format}, Quality: ${quality}`);

        DownloadYoutube(url, format, quality).then((result) => {
            console.log(result);
        }).catch((error) => {
            console.error(error);
        });
    };

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
                            <MenuItem value="360p">360p</MenuItem>
                            <MenuItem value="480p">480p</MenuItem>
                            <MenuItem value="720p">720p</MenuItem>
                            <MenuItem value="1080p">1080p</MenuItem>
                            <MenuItem value="4k">4k</MenuItem>
                        </TextField>
                    )}
                    <Button
                        fullWidth
                        variant="contained"
                        color="primary"
                        onClick={handleDownload}
                        style={{ marginTop: '16px' }}
                    >
                        Download
                    </Button>
                </Stack>
            </Box>
        </Container >
    );
};

export default YTDownloadPage;