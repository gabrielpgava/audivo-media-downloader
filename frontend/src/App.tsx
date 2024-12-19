import React, { useState } from 'react';
import AppBar from '@mui/material/AppBar';
import Toolbar from '@mui/material/Toolbar';
import Typography from '@mui/material/Typography';
import Button from '@mui/material/Button';
import IconButton from '@mui/material/IconButton';
import MenuIcon from '@mui/icons-material/Menu';
import { createTheme, ThemeProvider } from '@mui/material/styles';
import { CssBaseline } from '@mui/material';
import Container from '@mui/material/Container';
import Box from '@mui/material/Box';
import YTDownloadPage from './ytDownloadPage';
import AppleMusicDownloadPage from './AppleMusicDownloadPage';

const theme = createTheme();

const App: React.FC = () => {
    const [page, setPage] = useState('youtube');

    const renderPage = () => {
        switch (page) {
            case 'youtube':
                return <YTDownloadPage />
            case 'appleMusic':
                return <AppleMusicDownloadPage />
            default:
                return null;
        }
    };

    return (
        <ThemeProvider theme={theme}>
            <CssBaseline />
            <AppBar position="static">
                <Toolbar>
                    <Button color="inherit" onClick={() => setPage('youtube')}>YouTube</Button>
                    <Button color="inherit" onClick={() => setPage('appleMusic')}>Apple Music</Button>
                </Toolbar>
            </AppBar>
            <Container>
                <Box sx={{ my: 4 }}>
                    {renderPage()}
                </Box>
            </Container>
        </ThemeProvider >
    );
};

export default App;