import axios from "axios"
import AwaitLock from "await-lock"
import constants from '@/constants'
import store from '../store';  // Adjust the path to your store
import router from "../router";
const tokenRefreshLock = new AwaitLock()
let axiosInstance ;
let flagrAxiosInstance;

const API_URLS = constants.API_URLS
const API_URL = constants.API_URL
const SSO_URL = constants.SSO_URL

const ENVS = constants.ENVS
const ENVURLS = constants.ENVURLS


export function isLocalNetwork(hostname) {
	return (
		hostname.startsWith("localhost") ||
		hostname.startsWith("127.0.0.1") ||
		hostname.startsWith("192.168.") ||
		hostname.startsWith("10.0.") ||
		hostname.startsWith("0.0.0.0") ||
		hostname.endsWith(".local")
	)
}
  
export function determineEnv(hostname) {
	if (!hostname) {
	  console.warn("Couldn't determine host. Using production environment.")
	  return ENVS.DEV
	}
  
	if (isLocalNetwork(hostname) || hostname.includes("dev") || hostname.includes("demo")) {
	  return ENVS.DEV
	}
  
	if (hostname.includes("stage")) {
	  return ENVS.STAGE
	}

	if(hostname.includes("live")) {
		return ENVS.PROD
	}
  
	return ENVS.DEV
}

export const setupAxiosInstance = () => {
	let ssoURL = SSO_URL
	const env = determineEnv(window.location.host)
	ssoURL = ENVURLS[env].VUE_APP_SSO_API_URL 
	axiosInstance = axios.create({
		baseURL: ssoURL,
		timeout: 300000,
		headers: headerParams(),
	})
	setupAxiosInstanceInterceptors(axiosInstance)
	return axiosInstance
}

export const setupflagrAxiosInstance = () => {
	let apiurl = API_URL
	const env = determineEnv(window.location.host)
	apiurl = ENVURLS[env].VUE_APP_API_URL
	flagrAxiosInstance = axios.create({
		baseURL: apiurl,
		timeout: 300000
	})
	setupFlagrAxiosInstanceInterceptors(flagrAxiosInstance)
	return flagrAxiosInstance
}

export function setupAxiosInstanceInterceptors(axios){
	axios.interceptors.request.use((config) => {
		return config
	})
	axios.interceptors.response.use(
		(response) => {
			if (response.headers["x-access-token"] && response.headers["x-refresh-token"]) {
				const token = {
					"x-access-token": response.headers["x-access-token"] ,
					"x-refresh-token": response.headers["x-refresh-token"] ,
				}
				localStorage.setItem("tokens", JSON.stringify(token))
			}
			return response
		},
		async (error) => {
			if (error?.response?.status === 403) {
				logout()
			}
			return Promise.reject(error)
		},
	)
}

export function setupFlagrAxiosInstanceInterceptors(axios){
	axios.interceptors.request.use((config) => {
		return config
	})
	axios.interceptors.response.use(
		(response) => {
			if (response.headers["x-access-token"] && response.headers["x-refresh-token"]) {
				const token = {
					"x-access-token": response.headers["x-access-token"] ,
					"x-refresh-token": response.headers["x-refresh-token"] ,
				}
				localStorage.setItem("tokens", JSON.stringify(token))
			}
			return response
		},
		async (error) => {
			const originalRequest = error?.config
			if (error?.response?.status === 403) {
			  logout()
			}
			if (error?.response?.status === 401 && originalRequest?.["_retry"] !== true) {
			  const tokens = getToken()
			  if (tokens && originalRequest) {
				const originalRequestAccessToken = (originalRequest.headers["Authorization"])
				  ?.toString()
				  .split(" ")[1]
				await tokenRefreshLock.acquireAsync()
				try {
				  const newTokens = getToken()
				  originalRequest["_retry"] = true
				  if (newTokens && newTokens["x-access-token"] !== originalRequestAccessToken) {
					return axios(originalRequest)
				  }
				  // default workflow
				  await refreshToken({'x-refresh-token': tokens["x-refresh-token"]})
				  const latestToken = getToken()
				  originalRequest.headers["Authorization"] = "Bearer " + latestToken["x-access-token"];
				  const response = await axios(originalRequest)
				  return response
				} finally {
				  tokenRefreshLock.release()
				}
			  }
			}
			return Promise.reject(error)
		},
	)
}

const deviceType = ["web", "mweb", "ios-webview", "android-webview"]

export const detectDeviceType = () => {
	if (typeof window !== "undefined") {
		const details = navigator.userAgent
		const regexp = /android|iphone|kindle|ipad/i
		const isMobileDevice = regexp.test(details)
		if (isMobileDevice) {
			return deviceType[1]
		} else {
			return deviceType[0]
		}
	}
}

const headerParams = (default_headers = {}) => {
	const headers = default_headers
	headers["Content-Type"] = "application/json"
	headers["x-client-type"] = detectDeviceType() ?? "unknown"
	headers["accept"] = "application/json"
	if (typeof window !== "undefined") {
		let uuid = localStorage.getItem("uuid")
		if (!uuid) {
			uuid = generateUUID()
			localStorage?.setItem("uuid", uuid)
		}
		headers["x-device-id"] = uuid
	}
	return headers
}


export const generateUUID = () => {
	let d = new Date().getTime()
	let d2 = (typeof performance !== "undefined" && performance.now && performance.now() * 1000) || 0 //Time in microseconds since page-load or 0 if unsupported
	return "xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx".replace(/[xy]/g, function (c) {
		let r = Math.random() * 16 //random number between 0 and 16
		if (d > 0) {
			//Use timestamp until depleted
			r = (d + r) % 16 | 0
			d = Math.floor(d / 16)
		} else {
			//Use microseconds since page-load if supported
			r = (d2 + r) % 16 | 0
			d2 = Math.floor(d2 / 16)
		}
		return (c === "x" ? r : (r & 0x3) | 0x8).toString(16)
	})
}

export const getAxiosInstance = () => {
	if (!axiosInstance) {
		throw new Error("Axios instance not initialized")
	}
	const token = localStorage.getItem('tokens');

	// Try to parse the token and catch any potential errors
	if(token){
		try {
			const authToken = token ? JSON.parse(token) : null;
			if (authToken && authToken["x-access-token"]) {
				axiosInstance.defaults.headers["Authorization"] = "Bearer " + authToken["x-access-token"];
			}
		} catch (error) {
			console.error("Failed to parse token from localStorage", error);
		}
	} else {
		delete axiosInstance.defaults.headers["Authorization"]
	}

	// Set the Authorization header if the token exist

	return axiosInstance;
}

const getToken = () => {
	const token = localStorage.getItem('tokens');
	const authToken = token ? JSON.parse(token) : null;
	return authToken
}


export const getAxiosFlagrInstance = () => {
	if (!flagrAxiosInstance) {
		throw new Error("Axios instance not initialized")
	}
	const token = localStorage.getItem('tokens');

	// Try to parse the token and catch any potential errors
	if(token){
		try {
			const authToken = token ? JSON.parse(token) : null;
			if (authToken && authToken["x-access-token"]) {
				flagrAxiosInstance.defaults.headers["Authorization"] = "Bearer " + authToken["x-access-token"];
			}
		} catch (error) {
			console.error("Failed to parse token from localStorage", error);
		}
	} else {
		delete flagrAxiosInstance.defaults.headers["Authorization"]
	}

	// Set the Authorization header if the token exist

	return flagrAxiosInstance;
}


export const getGoogleRedirectionLink = async (postData) => {
    const response = await getAxiosInstance().get(
		`${`api/v1/auth/authenticate/googleAuthRedirectLink`}?uiRedirectUrl=${postData.uiRedirectUrl}`,
	)
	const redirectUrlResponse = response?.data?.data
	return redirectUrlResponse
}


export async function refreshToken(configHeader = {}) {
	const ssoInstance =  getAxiosInstance();
	const API_URL = `${API_URLS.REFRESH_TOKEN}`
	const response = await ssoInstance.get(API_URL, {
		headers: configHeader
	})
	if (response.status > 299) {
		throw response
	}
	return response?.data
}

export const googleAuthenticate = async (postData) => {	
	const reqPayload = JSON.stringify(postData)
	try {
		const response = await getAxiosInstance().post(`api/v1/auth/authenticate/google`, reqPayload)
		if (typeof window !== "undefined") {
			const token = {}
			token["x-access-token"] = response.headers["x-access-token"]
			token["x-refresh-token"] = response.headers["x-refresh-token"]
			localStorage.setItem("tokens", JSON.stringify(token))
		}
		return response?.data?.data
	} catch (error) {
		console.error(error)
		if (typeof window !== "undefined") {
			localStorage.removeItem("tokens")
		}
		throw error
	}
}

export function isAuthenticated() {
	return !!localStorage.getItem('tokens'); // Assuming token is saved in localStorage
}

export async function logout() {
	
	try {
		await getAxiosInstance().post(`api/v1/auth/logout`)
	} catch (error) {
		alert('logging you out');
	} finally {
		store.dispatch("reset")
		if (typeof window !== "undefined") {
			localStorage.removeItem("tokens")
			router.replace({ name: 'Login' })
		}
	}
}

export async function getUserDetails() {
	const response = await getAxiosInstance().get(`${API_URLS.USER_DETAILS}`)
	return response.data.data
}