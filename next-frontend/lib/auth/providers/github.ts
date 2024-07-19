import GithubProvider from 'next-auth/providers/github';


if(process.env.GITHUB_ID === "" || process.env.GITHUB_SECRET ==="") {
    throw new Error("GITHUB_ID and GITHUB_SECRET must be set in the environment variables")
}

export const githubProvider = GithubProvider({
    clientId: process.env.GITHUB_ID ?? '',
    clientSecret: process.env.GITHUB_SECRET ?? ''
})