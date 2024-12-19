import React, { useState } from 'react';
import { Container, TextField, Button, Typography, Box } from '@mui/material';
import Grid from '@mui/material/Grid2';
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
        <Container maxWidth="sm" style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', minHeight: '100vh' }}>
            <Typography variant="h4" gutterBottom color='black'>
                Apple Music Downloader
            </Typography>
            <Box component="form" noValidate autoComplete="off" style={{ width: '100%' }}>
                <Grid container spacing={2}>
                    <Grid>
                        <TextField
                            fullWidth
                            label="Apple Music URL"
                            variant="outlined"
                            margin="normal"
                            value={url}
                            onChange={(e) => setUrl(e.target.value)}
                        />
                    </Grid>
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

export default AppleMusicDownloadPage;