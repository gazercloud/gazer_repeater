import React, {useState} from 'react';
import PropTypes from 'prop-types';
import AppBar from '@material-ui/core/AppBar';
import CssBaseline from '@material-ui/core/CssBaseline';
import Divider from '@material-ui/core/Divider';
import Drawer from '@material-ui/core/Drawer';
import Hidden from '@material-ui/core/Hidden';
import IconButton from '@material-ui/core/IconButton';
import List from '@material-ui/core/List';
import ListItem from '@material-ui/core/ListItem';
import ListItemIcon from '@material-ui/core/ListItemIcon';
import ListItemText from '@material-ui/core/ListItemText';
import MenuIcon from '@material-ui/icons/Menu';
import Toolbar from '@material-ui/core/Toolbar';
import Typography from '@material-ui/core/Typography';
import {makeStyles, MuiThemeProvider, useTheme} from '@material-ui/core/styles';
import BlurOnIcon from '@material-ui/icons/BlurOn';
import InfoOutlinedIcon from '@material-ui/icons/InfoOutlined';
import PageAbout from "./Pages/PageAbout";
import Grid from "@material-ui/core/Grid";
import SignIn from "./Pages/SignIn";
import PageAccount from "./Pages/PageAccount";
import {createMuiTheme} from "@material-ui/core";
import PageNodes from "./Pages/PageNodes";
import PageRegistration from "./Pages/PageRegistration";
import PageConfirmationOk from "./Pages/PageConfirmationOk";
import PageConfirmationError from "./Pages/PageConfirmationError";
import PageChangePassword from "./Pages/PageChangePassword";
import PageResetPassword from "./Pages/PageResetPassword";
import PageRestorePassword from "./Pages/PageRestorePassword";
import PersonIcon from '@material-ui/icons/Person';

const drawerWidth = 240;


function getCookie(name) {
    let matches = document.cookie.match(new RegExp(
        "(?:^|; )" + name.replace(/([\.$?*|{}\(\)\[\]\\\/\+^])/g, '\\$1') + "=([^;]*)"
    ));
    return matches ? decodeURIComponent(matches[1]) : undefined;
}

const useStyles = makeStyles((theme) => ({
    root: {
        display: 'flex',
        color: '#D9D9D9',
        backgroundColor: '#121212'
    },
    drawer: {
        [theme.breakpoints.up('sm')]: {
            width: drawerWidth,
            flexShrink: 0
        },
    },
    appBar: {
        [theme.breakpoints.up('sm')]: {
            width: `calc(100% - ${drawerWidth}px)`,
            marginLeft: drawerWidth,
            backgroundColor: "#272727",
            color: "#CCC"
        },
    },
    menuButton: {
        marginRight: theme.spacing(2),
        [theme.breakpoints.up('sm')]: {
            display: 'none'
        },
    },
    // necessary for content to be below app bar
    toolbar: theme.mixins.toolbar,
    drawerPaper: {
        width: drawerWidth,
        color: "#EEEEEE",
        borderRight: '1px solid #850',
        backgroundColor: "#1E1E1E"
    },
    content: {
        flexGrow: 1,
        padding: theme.spacing(3),
    },
}));

function getHashVariable(variable) {
    const query = window.location.hash.substring(1);
    const vars = query.split('&');
    for (let i = 0; i < vars.length; i++) {
        const pair = vars[i].split('=');
        if (decodeURIComponent(pair[0]) === variable) {
            return decodeURIComponent(pair[1]);
        }
    }
    console.log('Query variable %s not found', variable);
}
/*
function getForm(variable) {
    const query = window.location.search.substring(1);
    const vars = query.split('&');
    for (let i = 0; i < vars.length; i++) {
        const pair = vars[i].split('=');
        if (decodeURIComponent(pair[0]) === variable) {
            return decodeURIComponent(pair[1]);
        }
    }
    console.log('Query variable %s not found', variable);
}
*/

const gotoLink = (link) => {
    window.location = link
}

function useForceUpdate(){
    const [value, setValue] = React.useState(0); // integer state
    return () => setValue(value => value + 1); // update the state to force render
}

function getWindow() {
    return window
}

function ResponsiveDrawer(props) {
    const { window } = props;
    const classes = useStyles();
    //const theme = useTheme();
    const [mobileOpen, setMobileOpen] = React.useState(false);
    const [redrawBool, setRedrawBool] = React.useState(false);
    const [title, setTitle] = React.useState(false);

    const forceUpdate = useForceUpdate();

    const navigate = (link) => {
        gotoLink(link)
        setMobileOpen(false)
        setRedrawBool(!redrawBool)
    }

    const updateTitle = (t) => {
        if (title !== t) {
            setTitle(t)
            document.title = t
        }
    }

    const handleDrawerToggle = () => {
        setMobileOpen(!mobileOpen);
    };

    const [firstRendering, setFirstRendering] = useState(true)
    if (firstRendering) {
        setFirstRendering(false)
    }

    getWindow().onhashchange = () => {
        setRedrawBool(!redrawBool)
    }

    const drawer = (
        <div>
            <Grid container className={classes.toolbar} direction="row" alignItems="center" alignContent="center" >

                <Grid container direction="row" alignItems="flex-start" alignContent="flex-start" style={{margin: "15px"}}>
                    <Grid item><a href="/"><img src="/mainicon32.png" style={{width: "48px"}}/></a></Grid>
                    <Grid item>
                        <Typography style={{marginLeft: "8px", marginBottom: "0px", fontSize: "14pt"}}>
                            Gazer Cloud
                        </Typography>
                        <Typography style={{marginLeft: "8px", fontSize: "10pt", color: "#BBB"}}>
                            control panel
                        </Typography>
                    </Grid>
                </Grid>

            </Grid>
            <Divider />
            <List>
                <ListItem button key="nodes" component="a" onClick={() => {navigate("#form=nodes")}}>
                    <ListItemIcon><BlurOnIcon style={{color:"#2278B5"}}/></ListItemIcon>
                    <ListItemText primary={"Nodes"} />
                </ListItem>
            </List>
            <Divider />
            <List>
                <ListItem button key="account" component="a" onClick={() => {navigate("#form=account")}}>
                    <ListItemIcon><PersonIcon style={{color:"#2278B5"}}/></ListItemIcon>
                    <ListItemText primary={"Account"} />
                </ListItem>
            </List>
            <Divider />
            <List>
                <ListItem button key="about" component="a" onClick={() => {navigate("#form=about")}}>
                    <ListItemIcon><InfoOutlinedIcon style={{color:"#2278B5"}}/></ListItemIcon>
                    <ListItemText primary={"Links"} />
                </ListItem>
            </List>
        </div>
    );

    const renderForm = () => {
        const form = getHashVariable("form")

        if (form === "nodes") {
            return (
                <PageNodes
                    OnNavigate={(addr)=> navigate(addr)}
                    OnTitleUpdate={(title) => updateTitle(title)}
                />
            )
        }

        if (form === "account") {
            return (
                <PageAccount OnTitleUpdate={(title) => updateTitle(title)} OnNavigate={(addr)=> navigate(addr)} OnNeedUpdate={()=>{ forceUpdate() }} />
            )
        }
        if (form === "change_password") {
            return (
                <PageChangePassword OnTitleUpdate={(title) => updateTitle(title)} OnNavigate={(addr)=> navigate(addr)} OnNeedUpdate={()=>{ forceUpdate() }} />
            )
        }
        if (form === "about") {
            return (
                <PageAbout OnNeedUpdate={()=>{ forceUpdate() }} />
            )
        }

        navigate("#form=nodes")

        return (
            <div>no form</div>
        )
    }

    const container = window !== undefined ? () => window().document.body : undefined;

    const th = createMuiTheme({
        palette: {
            type: 'dark',
            primary: {
                main: '#00A0E3'
            },
            secondary: {
                main: '#19BB4F'
            }
        }
    });

    if (getHashVariable("form") === "registration" && getCookie("session_token") === undefined)
    {
        return (
            <MuiThemeProvider theme={th}>
                <div className={classes.root}>
                    <PageRegistration OnNeedUpdate={ () => {
                        forceUpdate();
                    }}/>
                </div>
            </MuiThemeProvider>
        );
    }

    if (getHashVariable("form") === "reset_password" && getCookie("session_token") === undefined)
    {
        return (
            <MuiThemeProvider theme={th}>
                <div className={classes.root}>
                    <PageResetPassword OnNeedUpdate={ () => {
                        forceUpdate();
                    }}/>
                </div>
            </MuiThemeProvider>
        );
    }

    if (getHashVariable("form") === "restore_password" && getCookie("session_token") === undefined)
    {
        return (
            <MuiThemeProvider theme={th}>
                <div className={classes.root}>
                    <PageRestorePassword OnNeedUpdate={ () => {
                        forceUpdate();
                    }}/>
                </div>
            </MuiThemeProvider>
        );
    }

    if (getHashVariable("form") === "confirmation_ok")
    {
        return (
            <MuiThemeProvider theme={th}>
                <div className={classes.root}>
                    <PageConfirmationOk OnNeedUpdate={ () => {
                        forceUpdate();
                    }}/>
                </div>
            </MuiThemeProvider>
        );
    }

    if (getHashVariable("form") === "confirmation_error")
    {
        return (
            <MuiThemeProvider theme={th}>
                <div className={classes.root}>
                    <PageConfirmationError OnNeedUpdate={ () => {
                        forceUpdate();
                    }}/>
                </div>
            </MuiThemeProvider>
        );
    }

    if (getCookie("session_token") === undefined)
    {
        return (
            <MuiThemeProvider theme={th}>
                <div className={classes.root}>
                    <SignIn OnNeedUpdate={ () => {
                        forceUpdate();
                    }}/>
                </div>
            </MuiThemeProvider>
        );
    }

    return (
        <MuiThemeProvider theme={th}>
            <div className={classes.root} key='gazer-main'>
                <CssBaseline />
                <AppBar position="fixed" className={classes.appBar} style={{backgroundColor: '#1E1E1E'}}>
                    <Toolbar>
                        <IconButton
                            color="inherit"
                            aria-label="open drawer"
                            edge="start"
                            onClick={handleDrawerToggle}
                            className={classes.menuButton}
                        >
                            <MenuIcon />
                        </IconButton>
                        <Typography variant="h6" noWrap>
                            {title}
                        </Typography>
                    </Toolbar>
                </AppBar>
                <nav className={classes.drawer} aria-label="mailbox folders">
                    {/* The implementation can be swapped with js to avoid SEO duplication of links. */}
                    <Hidden smUp implementation="css">
                        <Drawer
                            container={container}
                            variant="temporary"
                            anchor={'left'}
                            open={mobileOpen}
                            onClose={handleDrawerToggle}
                            classes={{
                                paper: classes.drawerPaper,
                            }}
                            ModalProps={{
                                keepMounted: true, // Better open performance on mobile.
                            }}
                        >
                            {drawer}
                        </Drawer>
                    </Hidden>
                    <Hidden xsDown implementation="css">
                        <Drawer
                            classes={{
                                paper: classes.drawerPaper,
                            }}
                            variant="permanent"
                            open
                        >
                            {drawer}
                        </Drawer>
                    </Hidden>
                </nav>
                <main className={classes.content}>
                    <div className={classes.toolbar} />
                    {renderForm()}
                </main>
            </div>
        </MuiThemeProvider>
    );
}

ResponsiveDrawer.propTypes = {
    /**
     * Injected by the documentation to work in an iframe.
     * You won't need it on your project.
     */
    window: PropTypes.func,
};

export default ResponsiveDrawer;
