import React, { useState } from 'react';
import { Container, TextField, MenuItem, Button, Typography, Box } from '@mui/material';
import Grid from '@mui/material/Grid2';
import { DownloadYoutube, DownloadAppleMusic } from "../wailsjs/go/main/App";

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
        <Container maxWidth="sm" style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', minHeight: '100vh' }}>
            <Typography variant="h4" gutterBottom color='black'>
                YouTube Downloader
            </Typography>
            <Box component="form" noValidate autoComplete="off" style={{ width: '100%' }}>
                <Grid container spacing={2}>
                    <Grid>
                        <TextField
                            fullWidth
                            label="YouTube URL"
                            variant="outlined"
                            margin="normal"
                            value={url}
                            onChange={(e) => setUrl(e.target.value)}
                        />
                    </Grid>
                    <Grid>
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
                    </Grid>
                    {format === 'video' && (
                        <Grid>
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
                        </Grid>
                    )}
                    <Grid>
                        <Button
                            fullWidth
                            variant="contained"
                            color="primary"
                            onClick={handleDownload}
                            style={{ marginTop: '16px' }}
                        >
                            Download
                        </Button>
                    </Grid>
                </Grid>
            </Box>
        </Container>
    );
};

export default YTDownloadPage;