<?php

declare(strict_types=1);

namespace AnyVali;

enum UnknownKeyMode: string
{
    case Reject = 'reject';
    case Strip = 'strip';
    case Allow = 'allow';
}
