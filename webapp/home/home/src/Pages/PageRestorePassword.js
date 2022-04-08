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

export default function PageRestorePassword(props) {
    const classes = useStyles();
    const [userName, setUserName] = React.useState("")
    const [message, setMessage] = React.useState("")
    const [messageIsError, setMessageIsError] = React.useState(false)
    const [registrationIsOk, setRegistrationIsOk] = React.useState(false)

    const [firstRendering, setFirstRendering] = useState(true)
    if (firstRendering) {
        setUserName("")
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

    const requestRestorePassword = (token) => {
        setMessageIsError(false)
        setMessage("processing ...")
        let req = {
            "email": userName,
            "token": token
        }
        Request('s-restore-password', req)
            .then((res) => {
                if (res.status === 200) {
                    res.text().then(
                        (result) => {
                            try {
                                setMessageIsError(false)
                                setMessage("ok")
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
            });

    }

    const onClickRegister = (e) => {
        window.grecaptcha.ready(function() {
            window.grecaptcha.execute('6LdySEMbAAAAAAjLT9W3vZTpACerxYZmYmGRHDYP', {action: 'submit'}).then(function(token) {
                requestRestorePassword(token)
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
                        Enter your email below to receive an email with a link to change your password.
                    </Typography>
                    <form className={classes.form} noValidate>
                        <TextField
                            variant="outlined"
                            margin="normal"
                            required
                            fullWidth
                            id="email"
                            label="Email Address"
                            name="email"
                            autoComplete="email"
                            autoFocus
                            value={userName}
                            onChange={(event) => {
                                setUserName(event.target.value)
                            }}
                            onKeyPress={handleKey.bind(this)}
                        />
                        <Button
                            fullWidth
                            variant="contained"
                            color="primary"
                            onClick={onClickRegister}
                        >
                            Restore
                        </Button>
                    </form>
                </div>
                {messageIsError?
                    <div style={{color:'#F20', fontSize: '24pt'}}>
                        {message}
                    </div>:
                    <div style={{color:'#082', fontSize: '24pt'}}>
                        {message}
                    </div>
                }
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
                        Reset Your Password
                    </Typography>
                    <Typography>
                        You should receive an email shortly with information on how to reset your password.
                        If you do not receive an email within a few minutes, please try again.
                    </Typography>
                </div>
                <div style={{marginTop: "20px", color:'#082', fontSize: '24pt'}}>
                    Check your email: {userName}
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
