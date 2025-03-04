#!/bin/bash

MYDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
DAEMON_BINARIES=(kcn kpn ken kbn kscn kspn ksen)
BINARIES=(kgen homi)

set -e

function printUsage {
    echo "Usage: ${0} [-b] <arch> <target>"
    echo "         -b: use Kairos configuration"
    echo "     <arch>:  linux-386 | linux-amd64 | darwin-amd64 | windows-386 | windows-amd64"
    echo "   <target>:  kcn | kpn | ken | kbn | kscn | kspn | ksen | kgen | homi"
    echo ""
    echo "    ${0} linux-amd64 kcn"
    exit 1
}

# Parse options.
TESTNET=
while getopts "b" opt; do
    case ${opt} in
        b)
            echo "Using Kairos configuration..."
            TESTNET="-kairos"
            ;;
    esac
done
shift $((OPTIND -1))

# Parse subcommand.
SUBCOMMAND=$1
case "$SUBCOMMAND" in
	linux-386)
		PLATFORM_SUFFIX="linux-386"
		shift
		;;
	linux-amd64)
		PLATFORM_SUFFIX="linux-amd64"
		shift
		;;
	darwin-amd64)
		PLATFORM_SUFFIX="darwin-10.10-amd64"
		shift
		;;
	windows-386)
		PLATFORM_SUFFIX="windows-386"
		shift
		;;
	windows-amd64)
		PLATFORM_SUFFIX="windows-amd64"
		shift
		;;
	*)
		echo "Undefined architecture for packaging. Supported architectures: linux-386, linux-amd64, darwin-amd64, windows-386, windows-amd64"
		printUsage
		;;
esac

# Parse target
TARGET=$1
if [ -z "$TARGET" ]; then
    echo "specify target binary: ${DAEMON_BINARIES[*]} ${DAEMON[*]}"
    printUsage
fi

pushd $MYDIR/..

# Install trap to return PWD.
function finish {
  # Your cleanup code here
  popd
}
trap finish EXIT

KAIA_VERSION=$(go run build/rpm/main.go version)
KAIA_RELEASE_NUM=$(go run build/rpm/main.go release_num)
PACKAGE_SUFFIX="${KAIA_VERSION}-${KAIA_RELEASE_NUM}-${PLATFORM_SUFFIX}.tar.gz"

PACK_NAME=
KAIA_PACKAGE_NAME=
DAEMON=

# Search the target from DAEMON_BINARIES.
for b in ${DAEMON_BINARIES[*]}; do
    if [ "$TARGET" == "$b" ]; then
        PACK_NAME=${b}-${PLATFORM_SUFFIX}
        KAIA_PACKAGE_NAME="${b}${TESTNET}-${PACKAGE_SUFFIX}"
        DAEMON=1
    fi
done

# Search the target from BINARIES
for b in ${BINARIES[*]}; do
    if [ "$TARGET" == "$b" ]; then
        PACK_NAME=${b}-${PLATFORM_SUFFIX}
        KAIA_PACKAGE_NAME="${b}${TESTNET}-${PACKAGE_SUFFIX}"
    fi
done

if [ -z "$PACK_NAME" ]; then
    echo "Undefined binary name: $TARGET. Defined binaries: ${DAEMON_BINARIES[*]} ${DAEMON[*]}"

    printUsage
fi

# Copy the target binary
mkdir -p ${PACK_NAME}/bin
cp build/bin/${TARGET} ${PACK_NAME}/bin/${TARGET}

# Copy the configuration file and the daemon file.
if [ ! -z "$DAEMON" ]; then
    mkdir -p ${PACK_NAME}/conf
    CONF_FILE=build/packaging/linux/conf/${TARGET}d.conf
    if [ ! -z "$TESTNET" ]; then
        TESTNET_CONF_FILE=build/packaging/linux/conf/${TARGET}d_kairos.conf
        if [ -e "$TESTNET_CONF_FILE" ]; then
            CONF_FILE=$TESTNET_CONF_FILE
        else
            echo "Since $TESTNET_CONF_FILE does not exist, using $CONF_FILE."
        fi
    fi
	cp build/packaging/linux/bin/${TARGET}d ${PACK_NAME}/bin/
    cp $CONF_FILE ${PACK_NAME}/conf/
fi

# Compress!
mkdir -p packages
tar czf packages/$KAIA_PACKAGE_NAME $PACK_NAME
