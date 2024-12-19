import React, { useState } from 'react';
import { Container, TextField, Button, Typography, Box, Stack } from '@mui/material';
import { DownloadAppleMusic } from "../wailsjs/go/main/App";

const AppleMusicDownloadPage: React.FC = () => {
    const [url, setUrl] = useState('');
    const [format, setFormat] = useState('video');
    const [quality, setQuality] = useState('360p');

    const handleDownload = () => {
        console.log(`URL: ${url}, Format: ${format}, Quality: ${quality}`);

        DownloadAppleMusic(url, format, quality).then((result) => {
            console.log(result);
        }).catch((error) => {
            console.error(error);
        });
    };

    return (
        <Container maxWidth="sm" style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', minHeight: 'calc(100% - navbarpx)' }}>
            <Typography variant="h4" gutterBottom color='black'>
                Apple Music Downloader
            </Typography>
            <Box component="form" noValidate autoComplete="off" style={{ width: '100%' }}>
                <Stack>
                    <TextField
                        fullWidth
                        label="Apple Music URL"
                        variant="outlined"
                        margin="normal"
                        value={url}
                        onChange={(e) => setUrl(e.target.value)}
                    />
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
        </Container>
    );
};

export default AppleMusicDownloadPage;