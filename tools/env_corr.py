#!/usr/bin/env python3
"""Prefix-sum Pearson cross-correlation of 50 Hz RMS envelopes."""
import subprocess, sys, array, math

RATE, HOP = 8000, 160

def env(path):
    raw = subprocess.run(['ffmpeg','-hide_banner','-v','error','-i',path,
                          '-ac','1','-ar',str(RATE),'-f','s16le','-'],
                         capture_output=True, check=True).stdout
    a = array.array('h'); a.frombytes(raw[:len(raw)//2*2])
    return [math.sqrt(sum(v*v for v in a[i:i+HOP])/HOP)
            for i in range(0, len(a)-HOP, HOP)]

def prep(e):
    n=len(e); m=sum(e)/n
    z=[x-m for x in e]
    s=math.sqrt(sum(x*x for x in z)/n) or 1.0
    z=[x/s for x in z]
    P=[0.0]*(n+1); Q=[0.0]*(n+1); R=[0.0]*(n+1)
    for i,x in enumerate(z):
        P[i+1]=P[i]+x; Q[i+1]=Q[i]+x*x
    return z,P,Q,n

def corr_at(A, B, lag, L):
    az,Ap,Aq,an = A; bz,Bp,Bq,bn = B
    if lag >= 0:
        a0,b0 = lag, 0; L = min(an-lag, bn)
    else:
        a0,b0 = 0, -lag; L = min(an, bn+lag)
    if L < 200: return -2
    sa = Ap[a0+L]-Ap[a0]; sb = Bp[b0+L]-Bp[b0]
    saa= Aq[a0+L]-Aq[a0]; sbb= Bq[b0+L]-Bq[b0]
    # covariance via direct product is O(L); use overlap sum of products
    return None

def full(pathA, pathB, maxlag):
    A = prep(env(pathA)); B = prep(env(pathB))
    az,Ap,Aq,an = A; bz,Bp,Bq,bn = B
    best=[]
    for lag in range(-maxlag, maxlag):
        if lag>=0: a0,b0=lag,0; L=min(an-lag,bn)
        else: a0,b0=0,-lag; L=min(an,bn+lag)
        if L<500: continue
        # Pearson over overlap using prefix sums for means/var, product sum direct
        sa=(Ap[a0+L]-Ap[a0])/L; sb=(Bp[b0+L]-Bp[b0])/L
        saa=(Aq[a0+L]-Aq[a0])/L - sa*sa
        sbb=(Bq[b0+L]-Bq[b0])/L - sb*sb
        # product sum: block-window trick too slow; use windowed mean product
        prod=0.0
        for i in range(0, L, 8):
            j=min(i+8,L)
            for k in range(i,j):
                prod += az[a0+k]*bz[b0+k]
        prod/=L
        cov=prod-sa*sb
        d=math.sqrt(max(saa*sbb,1e-18))
        best.append((cov/d, lag))
    best.sort(reverse=True)
    return best[:5]

if __name__=='__main__':
    r = full(sys.argv[1], sys.argv[2], int(sys.argv[3]))
    print(f'{sys.argv[1].split("/")[-1]} vs {sys.argv[2].split("/")[-1]}')
    for c,l in r: print(f'   lag {l/50:+8.2f}s  r={c:+.4f}')
