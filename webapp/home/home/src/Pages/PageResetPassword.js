import React, {useState} from 'react';
import Avatar from '@material-ui/core/Avatar';
import Button from '@material-ui/core/Button';
import CssBaseline from '@material-ui/core/CssBaseline';
import TextField from '@material-ui/core/TextField';
import Link from '@material-ui/core/Link';
import Box from '@material-ui/core/Box';
import LockOutlinedIcon from '@material-ui/icons/LockOutlined';
import Typography from '@material-ui/core/Typography';
import { makeStyles } from '@material-ui/core/styles';
import Container from '@material-ui/core/Container';
import Request from "../request";
import {CircularProgress} from "@material-ui/core";

function RegCopyright() {
    return (
        <Typography variant="body2" color="textSecondary" align="center">
            {'Copyright © '}
            <Link color="inherit" href="https://gazer.cloud/">
                Gazer.Cloud
            </Link>{' '}
            {new Date().getFullYear()}
            {'.'}
        </Typography>
    );
}

const useStyles = makeStyles((theme) => ({
    paper: {
        marginTop: theme.spacing(8),
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
    },
    avatar: {
        margin: theme.spacing(1),
    },
    form: {
        width: '100%', // Fix IE 11 issue.
        marginTop: theme.spacing(1),
    },
    submit: {
        margin: theme.spacing(3, 0, 2),
    },
}));

export default function PageResetPassword(props) {
    const classes = useStyles();
    const [password1, setPassword1] = React.useState("")
    const [password2, setPassword2] = React.useState("")
    const [loading, setLoading] = React.useState(false)
    const [message, setMessage] = React.useState("")
    const [messageIsError, setMessageIsError] = React.useState(false)
    const [registrationIsOk, setRegistrationIsOk] = React.useState(false)


    const requestResetPassword= (token, key) => {
        if (password1 !== password2) {
            setMessageIsError(true)
            setMessage("passwords do not match")
            setLoading(false)
            return
        }

        setMessageIsError(false)
        setMessage("processing ...")
        let req = {
            "token": token,
            "password": password1,
            "key": key
        }
        Request('s-reset-password', req)
            .then((res) => {
                setLoading(false)
                if (res.status === 200) {
                    res.text().then(
                        (result) => {
                            try {
                                setMessageIsError(false)
                                setMessage("Your password has been reset.")
                                setRegistrationIsOk(true)
                                props.OnNeedUpdate()
                            } catch (e) {
                                setMessage("wrong server response")
                            }

                        }
                    )
                    return
                }
                if (res.status === 500) {
                    res.json().then(
                        (result) => {
                            setMessageIsError(true)
                            setMessage("Error: " + result.error)
                        }
                    );
                    return
                }

                res.text().then(
                    (result) => {
                        setMessage("Error " + res.status + ": " + result)
                    }
                )
            }).catch((e) => {
            setMessageIsError(true)
            setMessage("Error: " + e.message)
            setLoading(false)
        });

    }

    const [firstRendering, setFirstRendering] = useState(true)
    if (firstRendering) {
        setPassword1("")
        setPassword2("")
        setMessage("")
        setMessageIsError(false)
        setFirstRendering(false)
        {
            const script = document.createElement("script");
            script.src = "https://www.google.com/recaptcha/api.js?render=6LdySEMbAAAAAAjLT9W3vZTpACerxYZmYmGRHDYP";
            script.async = false;
            document.body.appendChild(script);
        }

        /*<script src="https://www.google.com/recaptcha/api.js?render=6LdySEMbAAAAAAjLT9W3vZTpACerxYZmYmGRHDYP"></script>*/
    }

    const getHashVariable =(variable) => {
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

    const onClickRegister = (e) => {
        setLoading(true)
        window.grecaptcha.ready(function() {
            window.grecaptcha.execute('6LdySEMbAAAAAAjLT9W3vZTpACerxYZmYmGRHDYP', {action: 'submit'}).then(function(token) {
                const resetPasswordKey = getHashVariable("key")
                requestResetPassword(token, resetPasswordKey)
            });
        });
    }

    const handleKey = (ev) => {
        if (ev.key === 'Enter') {
            ev.preventDefault();
            onClickRegister()
        }
    };

    const displayForm = () => {
        return (
            <Container component="main" maxWidth="xs">
                <CssBaseline />
                <div className={classes.paper}>
                    <Avatar className={classes.avatar}>
                        <LockOutlinedIcon />
                    </Avatar>
                    <Typography component="h1" variant="h5">
                        Reset Your Password
                    </Typography>
                    <Typography>
                        A secure alphanumeric password that contains 8+ characters is required.
                    </Typography>
                    <form className={classes.form} noValidate>
                        <TextField
                            variant="outlined"
                            margin="normal"
                            required
                            fullWidth
                            name="password1"
                            label="NEW PASSWORD"
                            type="password"
                            id="password1"
                            autoComplete="current-password"
                            value={password1}
                            onChange={(event) => {
                                setPassword1(event.target.value)
                            }}
                            onKeyPress={handleKey.bind(this)}
                        />
                        <TextField
                            variant="outlined"
                            margin="normal"
                            required
                            fullWidth
                            name="password2"
                            label="CONFIRM NEW PASSWORD"
                            type="password"
                            id="password2"
                            autoComplete="current-password"
                            value={password2}
                            onChange={(event) => {
                                setPassword2(event.target.value)
                            }}
                            onKeyPress={handleKey.bind(this)}
                        />

                        <Button
                            fullWidth
                            variant="contained"
                            color="primary"
                            onClick={onClickRegister}
                            disabled={loading}
                        >
                            RESET
                        </Button>

                        <Button
                            fullWidth
                            variant="outlined"
                            color="secondary"
                            style={{marginTop: "10px"}}
                            onClick={()=> {
                                window.location = "https://home.gazer.cloud/"
                            }}
                        >
                            CANCEL
                        </Button>
                    </form>
                </div>
                <div>
                {messageIsError?
                    <div style={{color:'#F20', fontSize: '24pt'}}>
                        <div>{message}</div>
                    </div>:
                    <div style={{color:'#082', fontSize: '24pt'}}>
                        <div>{message}</div>
                    </div>
                }
                </div>

                <div>{loading && <CircularProgress style={{margin: "0 auto"}} />}</div>

                <Box mt={8}>
                    <RegCopyright />
                </Box>
            </Container>
        )
    }

    const displayOK = () => {
        return (
            <Container component="main" maxWidth="xs">
                <CssBaseline />
                <div className={classes.paper}>
                    <Avatar className={classes.avatar}>
                        <LockOutlinedIcon />
                    </Avatar>
                    <Typography component="h1" variant="h5">
                        Reset password
                    </Typography>
                </div>
                <div style={{marginTop: "50px", color:'#082', fontSize: '24pt'}}>
                    Your password has been reset. Please use your new password to sign on again.
                </div>
                <div style={{marginTop: "20px", color:'#082', fontSize: '24pt'}}>
                    <a href="https://home.gazer.cloud/">LOGIN</a>.
                </div>
                <Box mt={8}>
                    <RegCopyright />
                </Box>
            </Container>
        )
    }

    if (registrationIsOk) {
        return displayOK();
    }

    return displayForm()
}
