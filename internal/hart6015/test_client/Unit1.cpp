//---------------------------------------------------------------------------

#include <vcl.h>
#include <stdio.h>

#pragma hdrstop
//---------------------------------------------------------------------------

HANDLE OpenPipe() {
    return CreateFileW( L"\\\\.\\pipe\\$TestHart$",
        GENERIC_READ | GENERIC_WRITE,
        FILE_SHARE_READ | FILE_SHARE_WRITE,
        NULL,
        OPEN_EXISTING,
        0,
        0);
}

bool ReadIntFromPipe(HANDLE hPipe, int &result ) {
    unsigned char b[4];
    DWORD readed_count;
    if ( !ReadFile(hPipe, b, 4, &readed_count, NULL) ){
        return false;
    }
    result = *((int *) b);
    return true;
}

bool ReadMessageFromPipe(HANDLE hPipe, int& level, AnsiString &text ) {

    if (!ReadIntFromPipe(hPipe, level ) ) {
        return false;
    }

    int strLen;
    if (!ReadIntFromPipe(hPipe, strLen ) ) {
        return false;
    }

    char *pStr = new char [strLen+1];
    DWORD readed_count;
    if ( !ReadFile(hPipe, pStr, strLen, &readed_count, NULL) ){
        printf("ReadFile failed");
        return false;
    }
    text = AnsiString(pStr, strLen );
    return true;
}



#pragma argsused
int main(int argc, char* argv[])
{
    std::system("MODE CON CP SELECT=1251");
    struct X { ~X() { getchar(); } } x;
    HANDLE hPipe = OpenPipe();
    if (hPipe == INVALID_HANDLE_VALUE){
        printf("hPipe == INVALID_HANDLE_VALUE");
        return 0;
    }
    AnsiString s;
    int level;

    while ( ReadMessageFromPipe(hPipe, level, s)) {
        AnsiString strLevel = "?";
        if(level==0){
            strLevel = "ERROR";
        }
        if(level==1){
            strLevel = "INFO";
        }
        if(level==2){
            strLevel = "DEBUG";
        }

        printf("%s|%s", strLevel.c_str(), s.c_str());
    }

    return 0;
}
//---------------------------------------------------------------------------
